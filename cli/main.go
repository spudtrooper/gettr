package cli

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/flags"
	"github.com/spudtrooper/goutil/formatstruct"
	goutilio "github.com/spudtrooper/goutil/io"
	"github.com/spudtrooper/goutil/or"
	"github.com/spudtrooper/goutil/parallel"
	"github.com/spudtrooper/goutil/sets"
)

var (
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

func isLimitExceeded(err error) bool {
	return strings.Contains(err.Error(), "E_METER_LIMIT_EXCEEDED")
}

func Main(ctx context.Context) error {
	app := makeApp()
	app.Init()

	f, err := model.MakeFactoryFromFlags(ctx)
	if err != nil {
		return err
	}
	client := f.Client()

	requireStringFlag := func(flag *string, name string) {
		if *flag == "" {
			log.Fatalf("--%s required", name)
		}
	}

	findFollowersWithExceptions := func(u *model.User, existing sets.StringSet) chan interface{} {
		followers := make(chan interface{})
		go func() {
			users, _ := u.Followers(ctx, model.UserFollowersMax(*max), model.UserFollowersMax(*threads), model.UserFollowersOffset(*offset))
			for u := range users {
				if ui, _ := u.UserInfo(ctx); ui.Username != "" {
					if existing[u.Username()] {
						log.Printf("skipping %s because we already follow them", u.Username())
					} else {
						followers <- u
					}
				}
			}
			close(followers)
		}()
		return followers
	}

	findFollowers := func(u *model.User) chan interface{} {
		return findFollowersWithExceptions(u, sets.StringSet{})
	}

	createPostWithImage := func(text, img string) (api.CreatePostInfo, error) {
		return client.CreatePost(text,
			api.CreatePostImages([]string{*postImage}),
			api.CreatePostDebug(*debug),
			api.CreatePostTitle(*postTitle),
			api.CreatePostPreviewImage(img),
			api.CreatePostPreviewSource(*postPreviewSource),
			api.CreatePostDescription(*dsc),
		)
	}

	createPost := func(text string) (api.CreatePostInfo, error) {
		return createPostWithImage(text, *postPreviewImage)
	}

	reply := func(postID, text string) (api.ReplyInfo, error) {
		return client.Reply(postID, text,
			api.ReplyDebug(*debug),
			api.ReplyTitle(*postTitle),
			api.ReplyPreviewImage(*postPreviewImage),
			api.ReplyPreviewSource(*postPreviewSource),
			api.ReplyDescription(*dsc),
		)
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

	mustFormatString := func(x interface{}) string {
		return fmt.Sprintf("<<<\n%s\n>>>", formatstruct.MustFormatString(x))
	}

	defaultUsername := func() string {
		if *other != "" {
			return *other
		}
		res := client.Username()
		log.Printf("defaulting to client user %q", res)
		return res
	}

	defaultUser := func() *model.User {
		username := defaultUsername()
		u := f.MakeUser(username)
		return u
	}

	self := func() *model.User {
		return f.Self()
	}

	maybePause := func() {
		if *pause != 0 {
			time.Sleep(*pause)
		}
	}

	shareAll := func(u *model.User) {
		followers := findFollowers(u)

		threads := or.Int(*threads, 20)
		parallel.ExecAndDrain(followers, threads, func(x interface{}) (interface{}, error) {
			f := x.(*model.User)
			posts, err := client.GetPosts(f.Username())
			if err != nil {
				return false, err
			}
			if len(posts) == 0 {
				return false, nil
			}
			post := posts[rand.Int()%len(posts)]
			log.Printf("%s trying to share: https://gettr.com/post/%s", f.Username(), post.ID)
			if err := client.SharePost(post.ID, *text, api.SharePostDebug(*debug)); err != nil {
				log.Printf("SharePost error: %v", err)
				if isLimitExceeded(err) {
					log.Fatalf("Limit exceeded: %v", err)
				}
			} else {
				log.Printf("shared https://gettr.com/post/%s from %s", post.ID, f.Username())
			}
			return true, nil
		})
	}

	replyToPost := func(post api.PostInfo) error {
		comments, err := client.GetComments(post.ID)
		if err != nil {
			return err
		}
		var comment string
		if len(comments) > 0 {
			comment = comments[rand.Int()%len(comments)].Text
		}
		if comment == "" {
			comment = "Nice work, homie"
		}
		log.Printf("trying to comment on: https://gettr.com/post/%s with %q", post.ID, comment)
		if _, err := reply(post.ID, comment); err != nil {
			log.Printf("Reply error: %v", err)
			if isLimitExceeded(err) {
				log.Fatalf("Limit exceeded: %v", err)
			}
		} else {
			log.Printf("commented on https://gettr.com/post/%s", post.ID)
		}
		return nil
	}

	replyAll := func(u *model.User) {
		followers := findFollowers(u)
		threads := or.Int(*threads, 20)
		parallel.ExecAndDrain(followers, threads, func(x interface{}) (interface{}, error) {
			f := x.(*model.User)
			posts, err := client.GetPosts(f.Username())
			if err != nil {
				return false, err
			}
			if len(posts) == 0 {
				return false, nil
			}
			post := posts[rand.Int()%len(posts)]
			if err := replyToPost(post); err != nil {
				return false, err
			}
			return true, nil
		})
	}

	app.Register("GetUserInfo", func() error {
		username := defaultUsername()
		info, err := client.GetUserInfo(username)
		if err != nil {
			return err
		}
		log.Printf("GetUserInfo: %s", mustFormatString(info))
		return nil
	})

	app.Register("GetPublicGlobals", func() error {
		info, err := client.GetPublicGlobals()
		if err != nil {
			return err
		}
		log.Printf("GetPublicGlobals: %s", mustFormatString(info))
		return nil
	})

	app.Register("GetSuggestions", func() error {
		info, err := client.GetSuggestions()
		if err != nil {
			return err
		}
		log.Printf("GetSuggestions: %s", mustFormatString(info))
		return nil
	})

	app.Register("GetPosts", func() error {
		username := defaultUsername()
		infos, err := client.GetPosts(username)
		if err != nil {
			return err
		}
		for i, info := range infos {
			log.Printf("GetPosts[%d]: %s", i, mustFormatString(info))
		}
		return nil
	})

	app.Register("Timeline", func() error {
		infos, err := client.Timeline()
		if err != nil {
			return err
		}
		for i, info := range infos {
			log.Printf("Timeline[%d]: %s", i, mustFormatString(info))
		}
		return nil
	})

	app.Register("LiveNow", func() error {
		infos, err := client.LiveNow()
		if err != nil {
			return err
		}
		for i, info := range infos {
			log.Printf("LiveNow[%d]: %s", i, mustFormatString(info))
		}
		return nil
	})

	app.Register("GetComments", func() error {
		requireStringFlag(postID, "post_id")
		infos, err := client.GetComments(*postID)
		if err != nil {
			return err
		}
		for i, info := range infos {
			log.Printf("Comments[%d]: %s", i, mustFormatString(info))
		}
		return nil
	})

	app.Register("GetPost", func() error {
		requireStringFlag(postID, "post_id")
		info, err := client.GetPost(*postID)
		if err != nil {
			return err
		}
		log.Printf("GetPost: %s", mustFormatString(info))
		return nil
	})

	app.Register("GetMuted", func() error {
		info, err := client.GetMuted()
		if err != nil {
			return err
		}
		log.Printf("GetMuted: %s", mustFormatString(info))
		return nil
	})

	app.Register("GetFollowings", func() error {
		username := defaultUsername()
		infos, err := client.GetFollowings(username, api.FollowingsOffset(*offset), api.FollowingsMax(*max))
		if err != nil {
			return err
		}
		for i, info := range infos {
			log.Printf("GetFollowings[%d]: %s", i, mustFormatString(info))
		}
		return nil
	})

	app.Register("GetAllFollowings", func() error {
		username := defaultUsername()
		infos, err := client.GetAllFollowings(username)
		if err != nil {
			return err
		}
		for i, info := range infos {
			log.Printf("GetAllFollowings[%d]: %s", i, mustFormatString(info))
		}
		for _, u := range infos {
			if err := client.Follow(u.Username); err != nil {
				return err
			}
			maybePause()
		}
		return nil
	})

	app.Register("GetFollowers", func() error {
		username := defaultUsername()
		infos, err := client.GetFollowers(username, api.FollowersOffset(*offset), api.FollowersMax(*max))
		if err != nil {
			return err
		}
		for i, info := range infos {
			log.Printf("GetFollowers[%d]: %s", i, mustFormatString(info))
		}
		return nil
	})

	app.Register("GetAllFollowers", func() error {
		username := defaultUsername()
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for i, u := range userInfos {
				log.Printf("users[%d][%d]: %v", offset, i, u)
				maybePause()
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
		return nil
	})

	app.Register("FollowAllCallback", func() error {
		requireStringFlag(other, "other")

		existingFollowers := sets.String(findFollowerUsernames(f.MakeUser(client.Username())))
		log.Printf("have %d existing followers", len(existingFollowers))

		username := *other
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following %d users[%d] of %s", len(userInfos), offset, username)
			for _, f := range userInfos {
				if existingFollowers[f.Username] {
					log.Printf("skipping %s because we already follow them", f.Username)
					continue
				}
				log.Printf("trying to follow %s", f.Username)
				if err := client.Follow(f.Username); err != nil {
					return err
				}
				log.Printf("followed %s", f.Username)
				maybePause()
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
		return nil
	})

	app.Register("FollowAll", func() error {
		requireStringFlag(other, "other")

		existingFollowers := sets.String(findFollowerUsernames(f.MakeUser(client.Username())))
		log.Printf("have %d existing followers", len(existingFollowers))
		u := f.MakeUser(*other)

		followers := findFollowersWithExceptions(u, existingFollowers)

		threads := or.Int(*threads, 20)
		parallel.ExecAndDrain(followers, threads, func(x interface{}) (interface{}, error) {
			f := x.(*model.User)
			log.Printf("trying to follow %s", f.Username())
			if err := client.Follow(f.Username()); err != nil {
				log.Printf("Follow error: %v", err)
				if isLimitExceeded(err) {
					log.Fatalf("Limit exceeded: %v", err)
				}
			} else {
				log.Printf("followed %s", f.Username())
			}
			maybePause()
			return nil, nil
		})
		return nil
	})

	app.Register("PrintAllFollowersCallback", func() error {
		username := defaultUsername()
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for _, f := range userInfos {
				fmt.Println(f.Username)
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
		return nil
	})

	app.Register("PrintAllFollowers", func() error {
		u := defaultUser()
		followers := findFollowers(u)
		i := 0
		for f := range followers {
			f := f.(*model.User)
			fmt.Printf("followers[%d] %s\n", i, f.Username())
			i++
		}
		return nil
	})

	app.Register("AllFollowersWebsites", func() error {
		u := defaultUser()
		followers := findFollowers(u)
		for f := range followers {
			f := f.(*model.User)
			website, err := f.Website(ctx)
			if err != nil {
				return err
			}
			if website != "" {
				fmt.Println(website)
			}
		}
		return nil
	})

	app.Register("PrintAllFollowingCallback", func() error {
		username := defaultUsername()
		if err := client.AllFollowings(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for _, f := range userInfos {
				fmt.Println(f.Username)
			}
			return nil
		}, api.AllFollowingsOffset(*offset)); err != nil {
			return err
		}
		return nil
	})

	app.Register("PrintAllFollowing", func() error {
		u := defaultUser()
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
		return nil
	})

	app.Register("AllFollowersFromFile", func() error {
		usernames, err := goutilio.StringsFromFile(*usernamesFile, goutilio.StringsFromFileSkipEmpty(true))
		if err != nil {
			return err
		}

		errs := make(chan error)
		out := make(chan string)
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
		return nil
	})

	app.Register("Persist", func() error {
		username := defaultUsername()
		user := f.MakeUser(username)
		if err := user.Persist(ctx,
			model.UserPersistMax(*max),
			model.UserPersistThreads(*threads),
			model.UserPersistForce(*force)); err != nil {
			return err
		}
		return nil
	})

	app.Register("Read", func() error {
		u := defaultUser()

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
		return nil
	})

	app.Register("PersistAll", func() error {
		u := defaultUser()

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
		return nil
	})

	app.Register("CreatePost", func() error {
		requireStringFlag(text, "text")
		info, err := createPost(*text)
		if err != nil {
			return err
		}
		log.Printf("CreatePost: %s", mustFormatString(info))
		return nil
	})

	app.Register("Reply", func() error {
		requireStringFlag(text, "text")
		requireStringFlag(postID, "postID")
		info, err := reply(*postID, *text)
		if err != nil {
			return err
		}
		log.Printf("Reply: %s", mustFormatString(info))
		return nil
	})

	app.Register("DeletePost", func() error {
		info, err := client.DeletePost(*postID)
		if err != nil {
			return err
		}
		log.Printf("DeletePost: %s", mustFormatString(info))
		return nil
	})

	app.Register("PersistInDB", func() error {
		u := defaultUser()
		if err := u.PersistInDB(ctx); err != nil {
			return err
		}
		return nil
	})

	app.Register("Upload", func() error {
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
		return nil
	})

	app.Register("CreatePostImage", func() error {
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
			res, err := createPostWithImage(*text, img)
			if err != nil {
				return err
			}
			log.Printf("CreatePost: %v", res)
		}
		return nil
	})

	app.Register("CreatePostCustomImage", func() error {
		requireStringFlag(uploadImage, "post_image")
		res, err := createPost(*text)
		if err != nil {
			return err
		}
		log.Printf("CreatePost: %v", res)
		return nil
	})

	app.Register("UpdateProfile", func() error {
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
		return nil
	})

	app.Register("LikeAll", func() error {
		u := defaultUser()

		followers := findFollowers(u)

		threads := or.Int(*threads, 20)
		parallel.ExecAndDrain(followers, threads, func(x interface{}) (interface{}, error) {
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
					log.Printf("LikePost error: %v", err)
					if isLimitExceeded(err) {
						log.Fatalf("Limit exceeded: %v", err)
					}
				} else {
					log.Printf("%s[%d] liked: https://gettr.com/post/%s", f.Username(), i, post.ID)
				}
			}
			return true, nil
		})
		return nil
	})

	app.Register("SharePostAll", func() error {
		u := self()
		shareAll(u)
		return nil
	})

	app.Register("SharePostFollowers", func() error {
		u := defaultUser()
		shareAll(u)
		return nil
	})

	app.Register("ReplyAll", func() error {
		u := self()
		replyAll(u)
		return nil
	})

	app.Register("ReplyFollowers", func() error {
		u := defaultUser()
		replyAll(u)
		return nil
	})

	app.Register("ReplyLiveNow", func() error {
		posts, err := client.LiveNow()
		if err != nil {
			return err
		}
		for _, post := range posts {
			replyToPost(post)
			maybePause()
		}
		return nil
	})

	app.Register("SharePost", func() error {
		requireStringFlag(postID, "post_id")
		requireStringFlag(text, "text")
		if err := client.SharePost(*postID, *text, api.SharePostDebug(*debug)); err != nil {
			return err
		}
		return nil
	})

	app.Register("Chat", func() error {
		requireStringFlag(postID, "post_id")
		requireStringFlag(text, "text")
		ok, err := client.Chat(*postID, *text)
		if err != nil {
			return err
		}
		log.Printf("Chat: %t", ok)
		return nil
	})

	app.Register("ChatLiveNow", func() error {
		requireStringFlag(text, "text")
		posts, err := client.LiveNow()
		if err != nil {
			return err
		}
		for _, post := range posts {
			log.Printf("chatting on %s", post.URI())
			ok, err := client.Chat(post.ID, *text)
			if err != nil {
				log.Printf("Chat error: %v", err)
				if isLimitExceeded(err) {
					panic("limit exceeded")
				}
			} else {
				log.Printf("Chat on %s: %t", post.URI(), ok)
			}
			maybePause()
		}
		return nil
	})

	app.Register("ChatThreads", func() error {
		requireStringFlag(postID, "post_id")
		requireStringFlag(text, "text")
		threads := or.Int(*threads, 200)
		parallel.DoTimes(threads, func() {
			ok, err := client.Chat(*postID, *text)
			if err != nil {
				log.Printf("error: %v", err)
				if isLimitExceeded(err) {
					panic("limit exceeded")
				}
			} else {
				log.Printf("Chat: %t", ok)
			}
		})
		return nil
	})

	app.Register("DeleteAll", func() error {
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
		return nil
	})

	app.Register("SearchPosts", func() error {
		requireStringFlag(query, "query")
		info, err := client.SearchPosts(*query, api.SearchMax(*max), api.SearchOffset(*offset), api.SearchDebug(*debug))
		if err != nil {
			return err
		}
		log.Printf("Search: %s", mustFormatString(info))
		for i, p := range info {
			log.Printf(" [%d] %s", i, p.URI())
		}
		return nil
	})

	app.Register("SearchUsers", func() error {
		requireStringFlag(query, "query")
		info, err := client.SearchUsers(*query, api.SearchMax(*max), api.SearchOffset(*offset), api.SearchDebug(*debug))
		if err != nil {
			return err
		}
		log.Printf("Search: %s", mustFormatString(info))
		for _, p := range info {
			log.Printf(" - %s", p.URI())
		}
		return nil
	})

	app.Register("AllPosts", func() error {
		posts, errors := client.AllPosts(*other, api.AllPostsMax(*max), api.AllPostsOffset(*offset), api.AllPostsThreads(*threads))
		parallel.WaitFor(func() {
			i := 0
			for ps := range posts {
				for _, p := range ps.Posts {
					log.Printf("post[%d @ %d]: %s", i, ps.Offset, p.URI())
					i++
				}
			}
		}, func() {
			for e := range errors {
				log.Printf("error: %v", e)
			}
		})
		return nil
	})

	app.Register("Help", func() error {
		app.ShowHelp()
		return nil
	})

	if err := app.Run(); err != nil {
		return err
	}

	return nil
}
