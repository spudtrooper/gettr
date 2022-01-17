package wordcloud

import (
	"context"
	"flag"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/hist"
)

var (
	other              = flag.String("other", "", "first user")
	max                = flag.Int("max", 0, "max to calls")
	threads            = flag.Int("threads", 0, "threads to calls")
	histLimit          = flag.Int("hist_limit", 0, "limit to # of histogram rows to print")
	followersThreads   = flag.Int("followers_threads", 300, "number of threads to use requesting --other's followers")
	maxFollowing       = flag.Int("max_following", 0, "max number of following to consider when building the histogram")
	printEveryUsername = flag.Bool("print_every_username", false, "print every username that we record")
)

func findFollowerUsernames(u *model.User) []string {
	fs, err := u.FollowersSync(api.AllFollowersMax(*max), api.AllFollowersMax(*threads))
	if err != nil {
		log.Printf("FollowersSync: igonoring: %v", err)
		return []string{}
	}
	var res []string
	for _, f := range fs {
		if f.Username() != "" {
			res = append(res, f.Username())
		}
	}
	return res
}

func findFollowingUsernames(u *model.User) []string {
	fs, err := u.FollowingSync(api.AllFollowingsMax(*max), api.AllFollowingsMax(*threads))
	if err != nil {
		log.Printf("FollowingSync: igonoring: %v", err)
		return []string{}
	}
	var res []string
	for _, f := range fs {
		if f.Username() != "" {
			res = append(res, f.Username())
		}
	}
	return res
}

var (
	nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

func Main(ctx context.Context) {
	if *other == "" {
		log.Fatalf("--other required")
	}

	factory, err := model.MakeFactoryFromFlags(ctx)
	check.Err(err)

	u := factory.MakeUser(*other)

	followers := make(chan *model.User)
	go func() {
		users, _ := u.Followers(ctx, model.UserFollowersMax(*max), model.UserFollowersMax(*threads))
		for u := range users {
			followers <- u
		}
		close(followers)
	}()

	b := hist.MakeHistogramChannelBuilder()
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < *followersThreads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for f := range followers {
					fs, err := f.GetFollowing(ctx, model.UserInfoDontRetry(true))
					if err != nil {
						log.Printf("ignoring followers error user for %s", f.Username())
						continue
					}
					if *maxFollowing > 0 && fs > *maxFollowing {
						continue
					}
					desc, err := f.Desc(ctx)
					if err != nil {
						log.Printf("ignoring description error user for %s", f.Username())
						f.MarkSkipped()
						continue
					}
					desc = strings.TrimSpace(desc)
					if desc == "" {
						continue
					}
					if *printEveryUsername {
						log.Printf("%s: %s", f.Username(), desc)
					}
					desc = strings.ToLower(desc)
					for _, w := range strings.Split(desc, " ") {
						s := w
						s = strings.TrimSpace(s)
						s = nonAlphaNum.ReplaceAllString(s, "")
						if s != "" {
							b.Add(s)
						}
					}
				}
			}()
		}
		wg.Wait()
		b.Close()
	}()

	h := b.Build(hist.MakeHistogramSortAsc(true))
	for i, p := range h.Pairs() {
		log.Printf("%d: %s", p.Value, p.Key)
		if *histLimit > 0 && i >= *histLimit {
			break
		}
	}
}
