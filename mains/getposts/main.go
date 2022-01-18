// Crawls a particular users's followers and imports all of their posts and their followers posts.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/gettr/todo"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/flags"
	"github.com/spudtrooper/goutil/or"
	"github.com/spudtrooper/goutil/parallel"
	"github.com/spudtrooper/goutil/ranges"
	"github.com/spudtrooper/goutil/sets"
	"github.com/spudtrooper/goutil/timing"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	other               = flags.String("other", "the other user")
	followersThreads    = flags.Int("follower_threads", "# of threads to calls for followers")
	followersMax        = flags.Int("followers_max", "max to calls for followers")
	postsThreads        = flags.Int("posts_threads", "# of threads to calls for posts")
	processThreads      = flags.Int("process_threads", "# of threads for processing the followers")
	postsMax            = flags.Int("posts_max", "max to calls for posts")
	forceFollowers      = flags.Bool("force_followers", "force to pull fresh followers")
	restart             = flags.Bool("restart", "when true we create the queue of followers")
	showMonitoringTitle = flags.Bool("show_monitoring_title", "show the monitoring title every so often")
)

const (
	allCollection        = "crawlFollowersQueue"
	doneCollection       = "crawlFollowersDone"
	maxOffsetsCollection = "crawlmaxOffsets"
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

	timing.SetLog(log.Printf)
	timing.GetOptions().Color = true

	type storedQueue struct {
		Username string
		Users    []string
	}

	type storedMaxOffset struct {
		Username string
		Offset   int
	}

	isNoDocError := func(err error) bool {
		return strings.Contains(fmt.Sprintf("%v", err), "mongo: no documents in result")
	}

	updateMaxOffset := func(username string, offset int) {
		filter := bson.D{{"username", username}}
		upsert := true
		opt := &options.FindOneAndUpdateOptions{
			Upsert: &upsert,
		}
		update := bson.D{
			{"$max", bson.D{
				{"offset", offset},
			}},
		}
		if res := db.Collection(maxOffsetsCollection).FindOneAndUpdate(ctx, filter, update, opt); res.Err() != nil {
			if skip := isNoDocError(res.Err()); !skip {
				todo.SkipErr("findMaxOff", res.Err())
			}
		}
	}

	findMaxOffset := func(username string) int {
		filter := bson.D{{"username", username}}
		var res storedMaxOffset
		if err := db.Collection(maxOffsetsCollection).FindOne(ctx, filter).Decode(&res); err != nil {
			if skip := isNoDocError(err); !skip {
				todo.SkipErr("findMaxOff", err)
			}
			return 0
		}
		return res.Offset
	}

	restartQueues := func() {
		timing.Push("restartQueues")
		defer timing.Pop()
		log.Printf("restarting queues")
		var users []string
		timing.Time("processUsers create usersToProcess", func() {
			usersToProcess := make(chan *model.User)
			go func() {
				followers, errs := other.Followers(ctx,
					model.UserFollowersForce(*forceFollowers),
					model.UserFollowersThreads(*followersThreads),
					model.UserFollowersThreads(*followersMax))
				parallel.WaitFor(func() {
					for u := range followers {
						usersToProcess <- u
					}
					usersToProcess <- other
				}, func() {
					for e := range errs {
						todo.SkipErr("Followers", e)
					}
				})
				close(usersToProcess)
			}()
			for u := range usersToProcess {
				users = append(users, u.Username())
			}
		})
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
		timing.Time("storeUsers(allCollection)", func() {
			storeUsers(allCollection, users)
		})
		timing.Time("storeUsers(doneCollection)", func() {
			storeUsers(doneCollection, []string{})
		})
		timing.Time("updateMaxOffsets", func() {
			usersCh := make(chan string)
			go func() {
				for _, u := range users {
					usersCh <- u
				}
				close(usersCh)
			}()
			var wg sync.WaitGroup
			go func() {
				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for u := range usersCh {
							updateMaxOffset(u, 0)
						}
					}()
				}
			}()
			wg.Wait()
		})
		// TODO: Remove done files
	}

	var processing bool
	process := func() {
		timing.Push("process")
		defer timing.Pop()

		usersToProcessCh := make(chan string)
		var numAll, numDone, numRemaining int
		{
			filter := bson.D{{"username", username}}

			var storedDone storedQueue
			var done sets.StringSet
			timing.Time("find(doneCollection)", func() {
				check.Err(db.Collection(doneCollection).FindOne(ctx, filter).Decode(&storedDone))
				done = sets.String(storedDone.Users)
			})

			var storedAll storedQueue
			var all []string
			timing.Time("find(allCollection)", func() {
				check.Err(db.Collection(allCollection).FindOne(ctx, filter).Decode(&storedAll))
				all = storedAll.Users
			})

			var remaining []string
			for _, u := range all {
				if !done[u] {
					remaining = append(remaining, u)
				}
			}
			numAll = len(all)
			numDone = len(done)
			numRemaining = len(remaining)
			cyan := func(i int) string {
				return color.New(color.FgCyan).Sprintf("%d", i)
			}
			log.Printf("found %s users to process", cyan(numAll))
			log.Printf("found %s users completed", cyan(numDone))
			log.Printf("found %s users remaining", cyan(numRemaining))
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

		var userCount, grandTotal, usersWithPosts int32
		processUser := func(user string) {
			start := findMaxOffset(user)
			posts, errors := f.Client().AllPosts(user,
				api.AllPostsThreads(postsThreads),
				api.AllPostsMax(*postsMax),
				api.AllPostsStart(start))
			atomic.AddInt32(&userCount, 1)
			var total int
			parallel.WaitFor(func() {
				for ps := range posts {
					total += len(ps.Posts)
					atomic.AddInt32(&grandTotal, int32(len(ps.Posts)))
					if len(ps.Posts) > 0 {
						if err := f.DB().AddPostInfos(ctx, user, ps.Posts); err != nil {
							todo.SkipErr("AddPosts", err)
						}
						updateMaxOffset(user, ps.Offset)
					}
				}
			}, func() {
				for e := range errors {
					log.Printf("error: %v", e)
				}
			})
			if total > 0 {
				atomic.AddInt32(&usersWithPosts, 1)
			}
		}

		start := time.Now()
		go func() {
			i := 0
			const tmplStart = "%s [%s/%s %s]: "
			const tmplEnd = "%s post(s) from %s user(s)"
			for processing {
				elapsed := time.Since(start)
				if *showMonitoringTitle && i%30 == 0 {
					log.Printf(tmplStart+"%s",
						color.New(color.FgHiYellow).Add(color.Underline).Sprintf("%20s", "Duration"),
						color.New(color.FgHiGreen).Add(color.Underline).Sprintf("%7s", "UsrCnt"),
						color.New(color.FgHiGreen).Add(color.Underline).Sprintf("%7s", "UsrTot"),
						color.New(color.FgHiCyan).Add(color.Underline).Sprintf("%7s", "% Done"),
						color.New(color.FgHiWhite).Add(color.Underline).Sprintf(tmplEnd, "# Posts", "# Users"),
					)
				}
				log.Printf(tmplStart+tmplEnd,
					color.New(color.FgYellow).Sprintf("%20v", elapsed),
					color.New(color.FgGreen).Sprintf("%7d", userCount),
					color.New(color.FgGreen).Sprintf("%7d", numRemaining),
					color.New(color.FgCyan).Sprintf("%6.2f%%", 100.0*float64(userCount)/float64(numRemaining)),
					color.New(color.FgHiWhite).Sprintf("%7d", grandTotal),
					color.New(color.FgHiWhite).Sprintf("%7d", usersWithPosts),
				)
				time.Sleep(3 * time.Second)
				i++
			}
		}()

		var wg sync.WaitGroup
		ranges.LoopTo(processThreads, func(i int) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for u := range usersToProcessCh {
					processUser(u)
					go func(u string) {
						markDone(u)
					}(u)
				}
			}()
		})
		wg.Wait()
	}

	if *restart {
		restartQueues()
	}

	processing = true
	process()
	processing = false
}

func main() {
	flag.Parse()
	crawl(context.Background())
}
