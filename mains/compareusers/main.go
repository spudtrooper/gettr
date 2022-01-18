package compareusers

import (
	"context"
	"flag"
	"log"

	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	a       = flag.String("a", "", "first user")
	b       = flag.String("b", "", "second user")
	max     = flag.Int("max", 0, "max to calls")
	threads = flag.Int("threads", 0, "threads to calls")
)

func findFollowers(ctx context.Context, u *model.User) []string {
	c := make(chan *model.User)
	go func() {
		users, _ := u.Followers(ctx, model.UserFollowersMax(*max), model.UserFollowersMax(*threads))
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

func compareUsers(ctx context.Context, a, b *model.User, factory model.Factory) {
	followersA := findFollowers(ctx, a)
	followersB := findFollowers(ctx, b)

	log.Printf("# folowersA: %d", len(followersA))
	log.Printf("# folowersB: %d", len(followersB))

	u := union(followersA, followersB)
	log.Printf("union: %d", len(u))

	d := difference(followersA, followersB)
	log.Printf("difference: %d", len(d))

	for i, un := range u {
		u := factory.MakeUser(un)
		log.Printf("union[%d]: %s", i, u.MustDebugString(ctx))
	}
}

func Main(ctx context.Context) {
	if *a == "" {
		log.Fatalf("--a required")
	}
	if *b == "" {
		log.Fatalf("--b required")
	}
	factory, err := model.MakeFactoryFromFlags(ctx)
	check.Err(err)
	userA := factory.MakeUser(*a)
	userB := factory.MakeUser(*b)
	compareUsers(ctx, userA, userB, factory)
}
