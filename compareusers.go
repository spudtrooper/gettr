package main

import (
	"flag"
	"log"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	a       = flag.String("a", "", "first user")
	b       = flag.String("b", "", "second user")
	max     = flag.Int("max", 0, "max to calls")
	threads = flag.Int("threads", 0, "threads to calls")
)

func findFollowers(u *model.User) []string {
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

// https://stackoverflow.com/questions/19374219/how-to-find-the-difference-between-two-slices-of-strings
func difference(a, b []string) []string {
	helper := func(a, b []string, m map[string]bool) {
		mb := make(map[string]struct{}, len(b))
		for _, x := range b {
			mb[x] = struct{}{}
		}
		for _, x := range a {
			if _, found := mb[x]; !found {
				m[x] = true
			}
		}
	}
	m := map[string]bool{}
	helper(b, a, m)
	log.Printf("after 2: %d", len(m))
	helper(a, b, m)
	log.Printf("after 1: %d", len(m))
	var res []string
	for s := range m {
		res = append(res, s)
	}
	return res
}

func union(a, b []string) []string {
	var res []string
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	for _, x := range a {
		if _, found := mb[x]; found {
			res = append(res, x)
		}
	}
	return res
}

func compareUsers(a, b *model.User, factory model.Factory) {
	followersA := findFollowers(a)
	followersB := findFollowers(b)

	log.Printf("# folowersA: %d", len(followersA))
	log.Printf("# folowersB: %d", len(followersB))

	u := union(followersA, followersB)
	log.Printf("union: %d", len(u))

	d := difference(followersA, followersB)
	log.Printf("difference: %d", len(d))

	for i, un := range u {
		u := factory.MakeUser(un)
		log.Printf("union[%d]: %s", i, u.MustDebugString())
	}
}

func realMain() {
	if *user == "" {
		log.Fatalf("--a required")
	}
	if *b == "" {
		log.Fatalf("--b required")
	}
	factory, err := model.MakeFactoryFromFlags()
	check.Err(err)
	userA := factory.MakeUser(*user)
	userB := factory.MakeUser(*b)
	compareUsers(userA, userB, factory)
}

func main() {
	flag.Parse()
	realMain()
}
