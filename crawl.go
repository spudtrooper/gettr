// Crawls a particular users's followers and imports all of their posts and their followers posts.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/gettr/todo"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/flags"
	"github.com/spudtrooper/goutil/or"
	"github.com/spudtrooper/goutil/parallel"
	"github.com/spudtrooper/goutil/ranges"
	"github.com/spudtrooper/goutil/sets"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	other            = flags.String("other", "the other user")
	followersThreads = flags.Int("follower_threads", "# of threads to calls for followers")
	followersMax     = flags.Int("followers_max", "max to calls for followers")
	postsThreads     = flags.Int("posts_threads", "# of threads to calls for posts")
	processThreads   = flags.Int("process_threads", "# of threads for processing the followers")
	postsMax         = flags.Int("posts_max", "max to calls for posts")
	forceFollowers   = flags.Bool("force_followers", "force to pull fresh followers")
	restart          = flags.Bool("restart", "when true we create the queue of followers")
)

const (
	allCollection  = "crawlFollowersQueue"
	doneCollection = "crawlFollowersDone"
)

func connect(ctx context.Context) (*mongo.Database, error) {
	const port = 27017
	const dbName = "gettrwork"

	uri := fmt.Sprintf("mongodb://localhost:%d", port)
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client.Database(dbName), nil
}

// For each user we maining two lists:
//   - allCollection: all the users to process; this never is mutated
//     after being filled.
//   - doneCollection: all the users that have been processed; we remove
//     a user after processing them.
//
// To restart set the --restart flag.
//
// For each user followed by --other, we:
//   - Find all posts and add these posts to the gettr DB `posts` table.
//   - Remove user from doneCollection
func crawl(ctx context.Context) {
	f, err := model.MakeFactoryFromFlags(ctx)
	check.Err(err)
	username := *other
	other := f.MakeUser(username)

	db, err := connect(ctx)
	check.Err(err)

	type storedQueue struct {
		Username string
		Users    []string
	}

	restartQueues := func() {
		usersToProcess := make(chan *model.User)
		go func() {
			followers, _ := other.Followers(ctx,
				model.UserFollowersForce(*forceFollowers),
				model.UserFollowersThreads(*followersThreads),
				model.UserFollowersThreads(*followersMax))
			for u := range followers {
				usersToProcess <- u
			}
			usersToProcess <- other
			close(usersToProcess)
		}()
		var users []string
		for u := range usersToProcess {
			users = append(users, u.Username())
		}
		storeUsers := func(collection string, users []string) {
			filter := bson.D{{"username", username}}
			db.Collection(collection).DeleteMany(ctx, filter)
			stored := storedQueue{
				Username: username,
				Users:    users,
			}
			if _, err := db.Collection(collection).InsertOne(ctx, stored); err != nil {
				check.Err(err)
			}
		}
		storeUsers(allCollection, users)
		storeUsers(doneCollection, []string{})
	}

	process := func() {
		usersToProcessCh := make(chan string)
		var numAll, numDone, numRemaining int
		{
			filter := bson.D{{"username", username}}

			var storedDone storedQueue
			check.Err(db.Collection(doneCollection).FindOne(ctx, filter).Decode(&storedDone))
			done := sets.String(storedDone.Users)

			var storedAll storedQueue
			check.Err(db.Collection(allCollection).FindOne(ctx, filter).Decode(&storedAll))
			all := storedAll.Users

			var remaining []string
			for _, u := range all {
				if !done[u] {
					remaining = append(remaining, u)
				}
			}
			numAll = len(all)
			numDone = len(done)
			numRemaining = len(remaining)
			log.Printf("found %d users to process", numAll)
			log.Printf("found %d users completed", numDone)
			log.Printf("found %d users remaining", numRemaining)
			go func() {
				for _, u := range remaining {
					usersToProcessCh <- u
				}
				close(usersToProcessCh)
			}()
		}

		var doneMu sync.Mutex
		markDone := func(user string) {
			doneMu.Lock()
			defer doneMu.Unlock()
			filter := bson.D{{"username", username}}
			var stored storedQueue
			check.Err(db.Collection(doneCollection).FindOne(ctx, filter).Decode(&stored))
			after := append(stored.Users, user)
			// TODO: What's the right way to update an array in mongodb?
			update := bson.D{
				{"$set", bson.D{
					{"users", after},
				}},
			}
			_, err := db.Collection(doneCollection).UpdateOne(ctx, filter, update)
			check.Err(err)
		}

		postsThreads := or.Int(*postsThreads, 2)
		log.Printf("getting posts with %d threads", postsThreads)
		processThreads := or.Int(*processThreads, 100)
		log.Printf("processing followers with %d threads", processThreads)

		var userCount, grandTotal int32
		processUser := func(user string) {
			posts, errors := f.Client().AllPosts(user,
				api.AllPostsThreads(postsThreads),
				api.AllPostsMax(*postsMax))
			log.Printf("processUser [%d/%d (%0.2f%%)]: %s",
				userCount, numRemaining, 100.0*float64(userCount)/float64(numRemaining), user)
			atomic.AddInt32(&userCount, 1)
			var total int
			parallel.WaitFor(func() {
				for ps := range posts {
					if err := f.DB().AddPostInfos(ctx, user, ps.Posts); err != nil {
						todo.SkipErr("AddPosts", err)
						continue
					}
					before := total
					total += len(ps.Posts)
					atomic.AddInt32(&grandTotal, int32(len(ps.Posts)))
					log.Printf("processUser [%d/%d (%0.2f%%)]: %s: +%d => %d posts (%d)",
						userCount, numRemaining, 100.0*float64(userCount)/float64(numRemaining),
						user, before, total, grandTotal)
				}
			}, func() {
				for e := range errors {
					log.Printf("error: %v", e)
				}
			})
			log.Printf("processUser processed %d posts for %s, grand total: %d", total, user, grandTotal)
		}

		var wg sync.WaitGroup
		ranges.LoopTo(processThreads, func(i int) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for u := range usersToProcessCh {
					processUser(u)
					markDone(u)
				}
			}()
		})
		wg.Wait()
	}

	if *restart {
		restartQueues()
	}
	process()
}

func main() {
	flag.Parse()
	crawl(context.Background())
}
