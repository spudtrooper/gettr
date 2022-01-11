package main

import (
	"flag"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	other     = flag.String("other", "", "first user")
	max       = flag.Int("max", 0, "max to calls")
	threads   = flag.Int("threads", 0, "threads to calls")
	limit     = flag.Int("limit", 0, "limit to # of followers to inspect")
	histLimit = flag.Int("hist_limit", 0, "limit to # of histogram rows to print")
)

func findFollowers(u *model.User) []*model.User {
	c := make(chan *model.User)
	go func() {
		users, _ := u.Followers(api.AllFollowersMax(*max), api.AllFollowersMax(*threads))
		for u := range users {
			c <- u
		}
		close(c)
	}()
	var res []*model.User
	for u := range c {
		res = append(res, u)
	}
	return res
}

func findFollowerUsernames(u *model.User) []string {
	c := make(chan *model.User)
	go func() {
		users, _ := u.Followers(api.AllFollowersMax(*max), api.AllFollowersMax(*threads))
		for u := range users {
			c <- u
		}
		close(c)
	}()
	var res []string
	for u := range c {
		res = append(res, u.Username())
	}
	return res
}

// https://stackoverflow.com/questions/18695346/how-can-i-sort-a-mapstringint-by-its-values
func sortMap(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func realMain() {
	if *other == "" {
		log.Fatalf("--other required")
	}

	client, err := api.MakeClientFromFlags()
	check.Err(err)
	cache, err := model.MakeCacheFromFlags()
	check.Err(err)
	factory := model.MakeFactory(cache, client)

	u := factory.MakeUser(*other)

	followers := make(chan *model.User)
	go func() {
		users, _ := u.Followers(api.AllFollowersMax(*max), api.AllFollowersMax(*threads))
		for u := range users {
			followers <- u
		}
		close(followers)
	}()

	keys := make(chan string)
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for f := range followers {
					userInfo, _ := f.UserInfo()
					if userInfo.Username == "" {
						log.Printf("ignoring empty user for %s", f.Username())
						continue
					}
					if userInfo.Followers() > 50 {
						continue
					}
					log.Printf("u: %s", f.Username())
					usernames := findFollowerUsernames(f)
					sort.Strings(usernames)
					key := strings.Join(usernames, ":")
					keys <- key
				}
			}()
		}
		log.Printf("done")
		wg.Wait()
		close(keys)
	}()

	hist := map[string]int{}
	for k := range keys {
		hist[k]++
	}

	sorted := sortMap(hist)
	for i, p := range sorted {
		log.Printf("%d: %s", p.Value, p.Key)
		if *histLimit > 0 && i >= *histLimit {
			break
		}
	}
}

func main() {
	flag.Parse()
	realMain()
}
