package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/flags"
	"github.com/spudtrooper/goutil/or"
	"github.com/spudtrooper/goutil/parallel"
	"github.com/spudtrooper/goutil/sets"
)

var (
	actions                = flag.String("actions", "", "comma-delimited list of calls to make")
	pause                  = flag.Duration("pause", 0, "pause amount between follows")
	offset                 = flag.Int("offset", 0, "offset for calls that take offsets")
	other                  = flag.String("other", "", "other username")
	usernamesFile          = flag.String("usernames_file", "", "file containing usernames")
	max                    = flag.Int("max", 0, "max to calls")
	threads                = flag.Int("threads", 0, "threads to calls")
	force                  = flag.Bool("force", false, "force things")
	text                   = flag.String("text", "", "text for posting")
	dsc                    = flag.String("dsc", "", "description for posting")
	postTitle              = flag.String("post_title", "", "title for posting")
	postID                 = flag.String("post_id", "", "post ID for deletion")
	uploadImage            = flags.String("upload_image", "image to upload")
	postImage              = flags.String("post_image", "image to post")
	postPreviewImage       = flags.String("post_preview_image", "preview image to post")
	postPreviewSource      = flags.String("post_preview_source", "preview source to post")
	profileDescription     = flags.String("profile_description", "profile description to update")
	profileLocation        = flags.String("profile_location", "profile location to update")
	profileWebsite         = flags.String("profile_website", "profile website to update")
	profileBackgroundImage = flags.String("profile_background_image", "profile background image to update")
	debug                  = flags.Bool("debug", "generic debug for some actions")
	query                  = flags.String("query", "query for search")
)

func realMain(ctx context.Context) error {
	factory, err := model.MakeFactoryFromFlags(ctx)
	if err != nil {
		return err
	}
	client := factory.Client()

	actionMap := map[string]bool{}
	if *actions != "" {
		for _, c := range strings.Split(*actions, ",") {
			actionMap[strings.ToLower(c)] = true
		}
	}
	for _, c := range flag.Args() {
		actionMap[strings.ToLower(c)] = true
	}
	shouldReturnedTrueOnce := false
	var possibleActions []string
	should := func(s string) bool {
		for k := range actionMap {
			if k == "all" {
				return true
			}
			if s == k {
				return actionMap[s]
			}
		}
		res := actionMap[strings.ToLower(s)]
		if res {
			shouldReturnedTrueOnce = true
		}
		possibleActions = append(possibleActions, s)
		return res
	}

	if len(actionMap) == 0 {
		return errors.Errorf("you need to specify at least one call")
	}

	requireStringFlag := func(flag *string, name string) {
		if *flag == "" {
			log.Fatalf("--%s required", name)
		}
	}

	if should("GetUserInfo") {
		info, err := client.GetUserInfo(*other)
		if err != nil {
			return err
		}
		log.Printf("GetUserInfo: %+v", info)
	}

	if should("GetPublicGlobals") {
		info, err := client.GetPublicGlobals()
		if err != nil {
			return err
		}
		log.Printf("GetPublicGlobals: %+v", info)
	}

	if should("GetSuggestions") {
		info, err := client.GetSuggestions()
		if err != nil {
			return err
		}
		log.Printf("GetSuggestions: %+v", info)
	}

	if should("GetPosts") {
		info, err := client.GetPosts(*other)
		if err != nil {
			return err
		}
		log.Printf("GetPosts: %+v", info)
	}

	if should("Timeline") {
		info, err := client.Timeline()
		if err != nil {
			return err
		}
		log.Printf("Timeline: %+v", info)
	}

	if should("GetComments") {
		info, err := client.GetComments("pmyaf4548d")
		if err != nil {
			return err
		}
		log.Printf("GetComments: %+v", info)
	}

	if should("GetPost") {
		info, err := client.GetPost("pmyaf4548d")
		if err != nil {
			return err
		}
		log.Printf("GetPost: %+v", info)
	}

	if should("GetMuted") {
		info, err := client.GetMuted()
		if err != nil {
			return err
		}
		log.Printf("GetMuted: %+v", info)
	}

	if should("GetFollowings") {
		info, err := client.GetFollowings(*other, api.FollowingsOffset(*offset), api.FollowingsMax(*max))
		if err != nil {
			return err
		}
		log.Printf("GetFollowings: %+v", info)
	}

	if should("GetAllFollowings") {
		info, err := client.GetAllFollowings(*other)
		if err != nil {
			return err
		}
		log.Printf("GetAllFollowings: %+v", info)
		for _, u := range info {
			if err := client.Follow(u.Username); err != nil {
				return err
			}
			if *pause > 0 {
				time.Sleep(*pause)
			}
		}
	}

	if should("GetFollowers") {
		info, err := client.GetFollowers(*other, api.FollowersOffset(*offset), api.FollowersMax(*max))
		if err != nil {
			return err
		}
		log.Printf("GetFollowers: %+v", info)
		for _, f := range info {
			log.Println(f.Username)
		}
	}

	if should("GetAllFollowers") {
		username := *other
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for i, u := range userInfos {
				log.Printf("users[%d][%d]: %v", offset, i, u)
				if *pause > 0 {
					time.Sleep(*pause)
				}
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
	}

	findFollowerUsernames := func(u *model.User) []string {
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

	if should("FollowAllCallback") {
		existingFollowersSet := sets.String(findFollowerUsernames(factory.MakeUser(client.Username())))
		log.Printf("have %d existing followers", len(existingFollowersSet))

		username := *other
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following %d users[%d] of %s", len(userInfos), offset, username)
			for _, f := range userInfos {
				if existingFollowersSet[f.Username] {
					log.Printf("skipping %s because we already follow them", f.Username)
					continue
				}
				log.Printf("trying to follow %s", f.Username)
				if err := client.Follow(f.Username); err != nil {
					return err
				}
				if *pause > 0 {
					time.Sleep(*pause)
				}
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
	}

	if should("FollowAll") {
		existingFollowersSet := sets.String(findFollowerUsernames(factory.MakeUser(client.Username())))
		log.Printf("have %d existing followers", len(existingFollowersSet))
		u := factory.MakeUser(*other)

		followers := make(chan *model.User)
		go func() {
			users, _ := u.Followers(ctx, model.UserFollowersMax(*max), model.UserFollowersMax(*threads), model.UserFollowersOffset(*offset))
			for u := range users {
				if ui, _ := u.UserInfo(ctx); ui.Username != "" {
					followers <- u
				}
			}
			close(followers)
		}()

		for f := range followers {
			if existingFollowersSet[f.Username()] {
				log.Printf("skipping %s because we already follow them", f.Username())
				continue
			}
			log.Printf("trying to follow %s", f.Username())
			if err := client.Follow(f.Username()); err != nil {
				return err
			}
			if *pause > 0 {
				time.Sleep(*pause)
			}
		}
	}

	if should("PrintAllFollowersCallback") {
		username := or.String(*other, client.Username())
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for _, f := range userInfos {
				fmt.Println(f.Username)
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
	}

	if should("PrintAllFollowers") {
		username := or.String(*other, client.Username())
		u := factory.MakeUser(username)

		followers := make(chan *model.User)
		go func() {
			users, _ := u.Followers(ctx, model.UserFollowersMax(*max), model.UserFollowersMax(*threads), model.UserFollowersOffset(*offset))
			for u := range users {
				if ui, _ := u.UserInfo(ctx); ui.Username != "" {
					followers <- u
				}
			}
			close(followers)
		}()

		i := 0
		for f := range followers {
			fmt.Printf("followers[%d] %s\n", i, f.Username())
			i++
		}
	}

	if should("PrintAllFollowingCallback") {
		username := or.String(*other, client.Username())
		if err := client.AllFollowings(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for _, f := range userInfos {
				fmt.Println(f.Username)
			}
			return nil
		}, api.AllFollowingsOffset(*offset)); err != nil {
			return err
		}
	}

	if should("PrintAllFollowing") {
		username := or.String(*other, client.Username())
		u := factory.MakeUser(username)

		following := make(chan *model.User)
		go func() {
			users, _ := u.Following(ctx, model.UserFollowingMax(*max), model.UserFollowingMax(*threads), model.UserFollowingOffset(*offset))
			for u := range users {
				ui, err := u.UserInfo(ctx)
				if err != nil {
					log.Printf("UserInfo: skipping error %v", err)
					continue
				}
				if ui.Username == "" {
					log.Printf("UserInfo: skipping user %s", u.Username())
					continue
				}
				following <- u
			}
			close(following)
		}()

		i := 0
		for f := range following {
			fmt.Printf("following[%d] %s\n", i, f.Username())
			i++
		}
	}

	if should("AllFollowersFromFile") {
		usernames := make(chan string)
		errs := make(chan error)
		out := make(chan string)

		f, err := os.Open(*usernamesFile)
		if err != nil {
			return err
		}
		defer f.Close()

		go func() {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if u := scanner.Text(); u != "" {
					usernames <- u
				}
			}
			close(usernames)
		}()

		go func() {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for u := range usernames {
						if err := client.Follow(u); err != nil {
							errs <- err
						} else {
							out <- u
						}
					}
				}()
			}
			wg.Wait()
			close(out)
			close(errs)
		}()

		for u := range out {
			log.Printf("done: %s", u)
		}
		for err := range errs {
			log.Fatalf("error: %v", err)
		}
	}

	if should("Persist") {
		user := factory.MakeUser(*other)
		if err := user.Persist(ctx, model.UserPersistMax(*max), model.UserPersistThreads(*threads), model.UserPersistForce(*force)); err != nil {
			return err
		}
	}

	if should("Read") {
		u := factory.MakeUser(*other)

		{
			c := make(chan *model.User)
			go func() {
				users, _ := u.Followers(ctx, model.UserFollowersMax(*max), model.UserFollowersMax(*threads))
				for u := range users {
					c <- u
				}
				close(c)
			}()

			i := 0
			for f := range c {
				log.Printf("followers[%d]: %s", i, f.Username())
				i++
			}
		}
		{
			c := make(chan *model.User)
			go func() {
				users, _ := u.Following(ctx, model.UserFollowingMax(*max), model.UserFollowingMax(*threads))
				for u := range users {
					c <- u
				}
				close(c)
			}()

			i := 0
			for f := range c {
				log.Printf("following[%d]: %s", i, f.Username())
				i++
			}
		}
	}

	if should("PersistAll") {
		u := factory.MakeUser(*other)

		{
			c := make(chan *model.User)
			go func() {
				users, _ := u.Followers(ctx, model.UserFollowersMax(*max), model.UserFollowersMax(*threads))
				for u := range users {
					c <- u
				}
				close(c)
			}()

			for f := range c {
				if err := f.Persist(ctx); err != nil {
					return err
				}
			}
		}
		{
			c := make(chan *model.User)
			go func() {
				users, _ := u.Following(ctx, model.UserFollowingMax(*max), model.UserFollowingMax(*threads))
				for u := range users {
					c <- u
				}
				close(c)
			}()

			for f := range c {
				if err := f.Persist(ctx); err != nil {
					return err
				}
			}
		}
	}

	if should("CreatePost") {
		info, err := client.CreatePost(*text,
			api.CreatePostDebug(*debug),
			api.CreatePostTitle(*postTitle),
			api.CreatePostPreviewImage(*postPreviewImage),
			api.CreatePostPreviewSource(*postPreviewSource),
			api.CreatePostDescription(*dsc),
		)
		if err != nil {
			return err
		}
		log.Printf("CreatePost: %+v", info)
	}

	if should("DeletePost") {
		info, err := client.DeletePost(*postID)
		if err != nil {
			return err
		}
		log.Printf("DeletePost: %+v", info)
	}

	if should("PersistInDB") {
		u := factory.MakeUser(*other)
		if err := u.PersistInDB(ctx); err != nil {
			return err
		}
	}

	if should("Upload") {
		requireStringFlag(uploadImage, "upload_image")
		var img string
		{
			res, err := client.Upload(*uploadImage)
			if err != nil {
				return err
			}
			log.Printf("Upload: %v", res)
			img = res.ORI
			if strings.HasPrefix(img, "/") {
				img = string(img[1:])
			}
		}
		{
			res, err := client.UpdateProfile(api.UpdateProfileIcon(img))
			if err != nil {
				return err
			}
			log.Printf("UpdateProfile: %v", res)
		}
	}

	if should("CreatePostImage") {
		requireStringFlag(uploadImage, "upload_image")
		var img string
		{
			res, err := client.Upload(*uploadImage)
			if err != nil {
				return err
			}
			log.Printf("Upload: %v", res)
			img = res.ORI
			if strings.HasPrefix(img, "/") {
				img = string(img[1:])
			}
		}
		{
			res, err := client.CreatePost(*text,
				api.CreatePostImages([]string{img}),
				api.CreatePostDebug(*debug),
				api.CreatePostTitle(*postTitle),
				api.CreatePostPreviewImage(*postPreviewImage),
				api.CreatePostPreviewSource(*postPreviewSource),
				api.CreatePostDescription(*dsc),
			)
			if err != nil {
				return err
			}
			log.Printf("CreatePost: %v", res)
		}
	}

	if should("CreatePostCustomImage") {
		requireStringFlag(uploadImage, "post_image")
		res, err := client.CreatePost(*text,
			api.CreatePostImages([]string{*postImage}),
			api.CreatePostDebug(*debug),
			api.CreatePostTitle(*postTitle),
			api.CreatePostPreviewImage(*postPreviewImage),
			api.CreatePostPreviewSource(*postPreviewSource),
			api.CreatePostDescription(*dsc),
		)
		if err != nil {
			return err
		}
		log.Printf("CreatePost: %v", res)
	}

	if should("UpdateProfile") {
		res, err := client.UpdateProfile(
			api.UpdateProfileBackgroundImage(*profileBackgroundImage),
			api.UpdateProfileDescription(*profileDescription),
			api.UpdateProfileLocation(*profileLocation),
			api.UpdateProfileWebsite(*profileWebsite),
		)
		if err != nil {
			return err
		}
		log.Printf("UpdateProfile: %v", res)
	}

	if should("LikeAll") {
		u := factory.Self()

		followers := make(chan interface{}) //*model.User)
		go func() {
			users, _ := u.Followers(ctx, model.UserFollowersMax(*max), model.UserFollowersMax(*threads), model.UserFollowersOffset(*offset))
			for u := range users {
				if ui, _ := u.UserInfo(ctx); ui.Username != "" {
					followers <- u
				}
			}
			close(followers)
		}()

		results, errors := parallel.Exec(followers, 10, func(x interface{}) (interface{}, error) {
			f := x.(*model.User)
			posts, err := client.GetPosts(f.Username())
			if err != nil {
				return false, err
			}
			if len(posts) == 0 {
				return false, nil
			}
			for i, post := range posts {
				log.Printf("%s[%d] trying to like: https://gettr.com/post/%s", f.Username(), i, post.ID)
				if err := client.LikePost(post.ID); err != nil {
					return false, err
				}
			}
			return true, nil
		})
		parallel.LazyDrain(results, errors)
	}

	if should("DeleteAll") {
		posts, err := client.GetPosts(client.Username())
		if err != nil {
			return err
		}
		for _, p := range posts {
			ok, err := client.DeletePost(p.ID)
			if err != nil {
				return err
			}
			log.Printf("deleted: %s -> %v", p.ID, ok)
		}
	}

	if should("SearchPosts") {
		requireStringFlag(query, "query")
		info, err := client.SearchPosts(*query, api.SearchMax(*max), api.SearchOffset(*offset), api.SearchDebug(*debug))
		if err != nil {
			return err
		}
		log.Printf("Search: %+v", info)
		for i, p := range info {
			log.Printf(" [%d] %s", i, p.URI())
		}
	}

	if should("SearchUsers") {
		requireStringFlag(query, "query")
		info, err := client.SearchUsers(*query, api.SearchMax(*max), api.SearchOffset(*offset), api.SearchDebug(*debug))
		if err != nil {
			return err
		}
		log.Printf("Search: %+v", info)
		for _, p := range info {
			log.Printf(" - %s", p.URI())
		}
	}

	if !shouldReturnedTrueOnce {
		var actions []string
		for s := range actionMap {
			actions = append(actions, fmt.Sprintf("%q", s))
		}
		msg := fmt.Sprintf("no valid actions in %v.\nThe possible actions are:\n", actions)
		sort.Strings(possibleActions)
		for _, s := range possibleActions {
			msg += " - " + s + "\n"
		}
		return errors.Errorf(msg)
	}

	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain(context.Background()))
}
