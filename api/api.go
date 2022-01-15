package api

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/log"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/flags"
	"github.com/spudtrooper/goutil/must"
	"github.com/spudtrooper/goutil/or"
)

var (
	clientStats = flags.Bool("client_stats", "Print client stats")
)

const (
	defaultMax     = 20
	defaultOffset  = 0
	defaultDir     = "fwd"
	defaultThreads = 200
	defaultStart   = 0
)

type Date interface {
	Time() (time.Time, error)
	String() string
}

type StringDate string

func (s StringDate) Time() (time.Time, error) {
	millis, err := strconv.Atoi(string(s))
	if err == nil {
		return time.Now(), err
	}
	return time.Unix(int64(millis/1000), int64(0)), nil
}

func (s StringDate) String() string {
	t, err := s.Time()
	check.Err(err)
	return fmt.Sprintf("%s@%v", string(s), t)
}

type StringInt string

func (s StringInt) Int() int {
	if s == "" {
		return 0
	}
	return must.Atoi(string(s))
}

type IntDate int

func (i IntDate) Time() (time.Time, error) {
	return time.Unix(int64(i/1000), int64(0)), nil
}

func (i IntDate) String() string {
	t, err := i.Time()
	check.Err(err)
	return fmt.Sprintf("%d@%v", i, t)
}

func (c *Client) GetUserInfo(username string) (UserInfo, error) {
	route := fmt.Sprintf("s/uinf/%s", username)
	var payload struct {
		Data UserInfo
	}
	if err := c.get(route, &payload); err != nil {
		return UserInfo{}, err
	}
	return payload.Data, nil
}

type BoolString string

func (b BoolString) Bool() (bool, error) {
	if b == "true" {
		return true, nil
	}
	if b == "false" {
		return false, nil
	}
	return false, errors.Errorf("invalid bool: %s", b)
}

type PublicGlobals struct {
	DisableQs     BoolString `json:"disable_qs"`
	DisableSignup BoolString `json:"disable_signup"`
	DisableSms    BoolString `json:"disable_sms"`
}

func (c *Client) GetPublicGlobals() (*PublicGlobals, error) {
	route := "u/public_globals"
	var payload struct {
		Globals PublicGlobals
	}
	if err := c.get(route, &payload); err != nil {
		return nil, err
	}
	return &payload.Globals, nil
}

type HtInfo struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	IconURL     string   `json:"iconUrl"`
	LiveURL     string   `json:"liveURL"`
	PostID      []string `json:"postId"`
	Title       string   `json:"title"`
	Topic       string   `json:"topic"`
}

func (c *Client) GetSuggestions(sOpts ...SuggestOption) ([]HtInfo, error) {
	opts := MakeSuggestOptions(sOpts...)
	max := or.Int(opts.Max(), 10)
	route := createRoute("s/hashtag/suggest", param{"max", max})
	type suggestions struct {
		HTInfo map[string]HtInfo `json:"htinfo"`
	}
	var payload struct {
		Aux suggestions `json:"aux"`
	}
	if err := c.get(route, &payload); err != nil {
		return nil, err
	}
	var res []HtInfo
	for _, s := range payload.Aux.HTInfo {
		res = append(res, s)
	}
	return res, nil
}

type PostInfo struct {
	CDate       IntDate `json:"cdate"`
	Update      IntDate `json:"update"`
	Text        string  `json:"txt"`
	TextLang    string  `json:"txt_lang"`
	Description string  `json:"dsc"`
	Type        string  `json:"_t"`
	ID          string  `json:"_id"`
	Comments    int     `json:"cm"`
	Likes       int     `json:"lkbpst"`
	Reposts     int     `json:"shbpst"`
}

func (c *Client) GetPosts(username string, pOpts ...PostsOption) ([]PostInfo, error) {
	opts := MakePostsOptions(pOpts...)
	offset := or.Int(opts.Offset(), defaultOffset)
	max := or.Int(opts.Max(), defaultMax)
	dir := or.String(opts.Dir(), defaultDir)
	incl := or.String(strings.Join(opts.Incl(), "|"), "posts|stats|userinfo|shared|liked")
	fp := or.String(opts.Fp(), "f_uo")
	route := createRoute(fmt.Sprintf("u/user/%s/posts", username),
		param{"offset", offset}, param{"max", max}, param{"dir", dir}, param{"incl", incl}, param{"fp", fp})
	type posts struct {
		Posts map[string]PostInfo `json:"post"`
	}
	var payload struct {
		Aux posts `json:"aux"`
	}
	if err := c.get(route, &payload); err != nil {
		return nil, err
	}
	var res []PostInfo
	for _, p := range payload.Aux.Posts {
		res = append(res, p)
	}
	return res, nil
}

type CommentInfo struct {
	CDate    IntDate  `json:"cdate"`
	Update   IntDate  `json:"update"`
	Text     string   `json:"txt"`
	TextLang string   `json:"txt_lang"`
	Type     string   `json:"_t"`
	ID       string   `json:"_id"`
	Hashtags []string `json:"htgs"`
	UID      string   `json:"uid"`
	PUID     string   `json:"puid"`
	PID      string   `json:"pid"`
}

func (c *Client) GetComments(post string, cOpts ...CommentsOption) ([]CommentInfo, error) {
	opts := MakeCommentsOptions(cOpts...)
	offset := or.Int(opts.Offset(), defaultOffset)
	max := or.Int(opts.Max(), defaultMax)
	dir := or.String(opts.Dir(), defaultDir)
	incl := or.String(strings.Join(opts.Incl(), "|"), "posts|stats|userinfo|shared|liked")
	route := createRoute(fmt.Sprintf("u/post/%s/comments", post),
		param{"offset", offset}, param{"max", max}, param{"dir", dir}, param{"incl", incl})
	type comments struct {
		Comments map[string]CommentInfo `json:"cmt"`
	}
	var payload struct {
		Aux comments `json:"aux"`
	}
	if err := c.get(route, &payload); err != nil {
		return nil, err
	}
	var res []CommentInfo
	for _, p := range payload.Aux.Comments {
		res = append(res, p)
	}
	return res, nil
}

type ShareInfo struct {
	Comments int `json:"cm"`
	Likes    int `json:"lkbpst"`
	Shares   int `json:"shbpst"`
}

type UserInfo struct {
	BGImg     string     `json:"bgimg"`
	Desc      string     `json:"dsc"`
	ICO       string     `json:"ico"`
	Infl      int        `json:"infl"`
	Lang      string     `json:"lang"`
	OUsername string     `json:"ousername"`
	Username  string     `json:"username"`
	Website   string     `json:"website"`
	TwtFlg    StringInt  `json:"twt_flg"`
	TwtFlw    StringInt  `json:"twt_flw"`
	Flg       StringInt  `json:"flg"`
	Flw       StringInt  `json:"flw"`
	CDate     StringDate `json:"cdate"`
	UDate     StringDate `json:"udate"`
	Type      string     `json:"_t"`
	ID        string     `json:"_id"`
	Nickname  string     `json:"nickname"`
	Status    string     `json:"status"`
}

func (u UserInfo) Following() int        { return u.Flw.Int() }
func (u UserInfo) Followers() int        { return u.Flg.Int() }
func (u UserInfo) TwitterFollowing() int { return u.TwtFlw.Int() }
func (u UserInfo) TwitterFollowers() int { return u.TwtFlg.Int() }

type UserInfos []UserInfo

type PostDetails struct {
	ShareInfo
	PostInfo
	UserInfos
}

func (c *Client) GetPost(post string, pOpts ...PostOption) (PostDetails, error) {
	opts := MakePostOptions(pOpts...)
	incl := or.String(strings.Join(opts.Incl(), "|"), "posts|stats|userinfo|shared|liked")
	route := createRoute(fmt.Sprintf("u/post/%s", post), param{"incl", incl})
	type aux struct {
		Share ShareInfo           `json:"s_pst"`
		Uinf  map[string]UserInfo `json:"uinf"`
	}
	var payload struct {
		Aux  aux      `json:"aux"`
		Data PostInfo `json:"data"`
	}
	if err := c.get(route, &payload); err != nil {
		return PostDetails{}, err
	}
	res := PostDetails{
		ShareInfo: payload.Aux.Share,
		PostInfo:  payload.Data,
	}
	for _, u := range payload.Aux.Uinf {
		res.UserInfos = append(res.UserInfos, u)
	}
	return res, nil
}

func (c *Client) GetMuted(mOpts ...MutedOption) (UserInfos, error) {
	opts := MakeMutedOptions(mOpts...)
	offset := or.Int(opts.Offset(), defaultOffset)
	max := or.Int(opts.Max(), defaultMax)
	incl := or.String(strings.Join(opts.Incl(), "|"), "userstats|userinfo")
	route := createRoute(fmt.Sprintf("u/user/%s/mutes", c.username), param{"offset", offset}, param{"max", max}, param{"incl", incl})
	type aux struct {
		Uinf map[string]UserInfo `json:"uinf"`
	}
	var payload struct {
		Aux aux `json:"aux"`
	}
	if err := c.get(route, &payload); err != nil {
		return nil, err
	}
	var res UserInfos
	for _, u := range payload.Aux.Uinf {
		res = append(res, u)
	}
	return res, nil
}

func (c *Client) GetFollowings(username string, fOpts ...FollowingsOption) (UserInfos, error) {
	opts := MakeFollowingsOptions(fOpts...)
	offset := or.Int(opts.Offset(), defaultOffset)
	max := or.Int(opts.Max(), defaultMax)
	incl := or.String(strings.Join(opts.Incl(), "|"), "userstats|userinfo")
	route := createRoute(fmt.Sprintf("u/user/%s/followings", username), param{"offset", offset}, param{"max", max}, param{"incl", incl})
	type aux struct {
		Uinf map[string]UserInfo `json:"uinf"`
	}
	var payload struct {
		Aux aux `json:"aux"`
	}
	if err := c.get(route, &payload); err != nil {
		return nil, err
	}
	var res UserInfos
	for _, u := range payload.Aux.Uinf {
		res = append(res, u)
	}
	return res, nil
}

func (c *Client) GetAllFollowings(username string, fOpts ...FollowingsOption) (UserInfos, error) {
	opts := MakeFollowingsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	var res UserInfos
	for offset := 0; ; offset += max {
		followings, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
		if err != nil {
			return nil, err
		}
		if len(followings) == 0 {
			break
		}
		res = append(res, followings...)
	}
	return res, nil
}

func (c *Client) AllFollowings(username string, f func(offset int, us UserInfos) error, fOpts ...AllFollowingsOption) error {
	opts := MakeAllFollowingsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	for offset := start; ; offset += max {
		followings, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
		if err != nil {
			return err
		}
		if len(followings) == 0 {
			break
		}
		if err := f(offset, followings); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) AllFollowingsParallel(username string, fOpts ...AllFollowingsOption) (chan UserInfo, chan error) {
	opts := MakeAllFollowingsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	threads := or.Int(opts.Threads(), defaultThreads)

	userInfos := make(chan UserInfo)
	offsets := make(chan int)
	errs := make(chan error)

	go func() {
		for offset := start; offset < math.MaxInt; offset += max {
			offsets <- offset
		}
		close(offsets)
	}()

	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for offset := range offsets {
					fs, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
					if err != nil {
						errs <- err
						break
					}
					if len(fs) == 0 {
						break
					}
					for _, u := range fs {
						userInfos <- u
					}
				}
			}()
		}
		wg.Wait()
		close(userInfos)
		close(errs)
	}()

	return userInfos, errs
}

func (c *Client) Follow(username string) error {
	route := createRoute(fmt.Sprintf("u/user/%s/follows/%s", c.username, username))
	if err := c.post(route, nil, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetFollowers(username string, fOpts ...FollowersOption) (UserInfos, error) {
	opts := MakeFollowersOptions(fOpts...)
	offset := or.Int(opts.Offset(), defaultOffset)
	max := or.Int(opts.Max(), defaultMax)
	incl := or.String(strings.Join(opts.Incl(), "|"), "userstats|userinfo")
	route := createRoute(fmt.Sprintf("u/user/%s/followers", username), param{"offset", offset}, param{"max", max}, param{"incl", incl})
	type aux struct {
		Uinf map[string]UserInfo `json:"uinf"`
	}
	var payload struct {
		Aux aux `json:"aux"`
	}
	if err := c.get(route, &payload); err != nil {
		return nil, err
	}
	var res UserInfos
	for _, u := range payload.Aux.Uinf {
		res = append(res, u)
	}
	return res, nil
}

type OffsetStrings struct {
	Offset  int
	Strings []string
}

type clientStatsCollector struct {
	tag   string
	start time.Time
	durs  []time.Duration
}

func makeClientStatsCollector(tag string) *clientStatsCollector {
	start := time.Now()
	return &clientStatsCollector{tag: tag, start: start}
}

func (c *clientStatsCollector) RecordAndPrint() {
	stop := time.Now()
	dur := stop.Sub(c.start)
	c.durs = append(c.durs, dur)
	median := func() int {
		if len(c.durs) == 0 {
			return 0
		}
		var durs []int
		for _, d := range c.durs {
			durs = append(durs, int(d))
		}
		sort.Ints(durs)
		mid := len(durs) - 1
		if mid%2 == 1 {
			return (durs[mid-1] + durs[mid]) / 2
		}
		return durs[mid]
	}
	mean := func() int {
		if len(c.durs) == 0 {
			return 0
		}
		var sum int
		for _, d := range c.durs {
			sum += int(d)
		}
		return sum / len(c.durs)
	}
	log.Printf("%s stats: samples=%s median=%s mean=%s",
		color.New(color.FgHiWhite).Sprintf("%s", c.tag),
		color.YellowString(fmt.Sprintf("%d", len(c.durs))),
		color.GreenString(fmt.Sprintf("%v", time.Duration(median()))),
		color.CyanString(fmt.Sprintf("%v", time.Duration(mean()))))
}

func (c *Client) AllFollowersParallel(username string, fOpts ...AllFollowersOption) (chan UserInfo, chan OffsetStrings, chan error) {
	opts := MakeAllFollowersOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	threads := or.Int(opts.Threads(), defaultThreads)

	offsets := make(chan int)
	go func() {
		for offset := start; offset < math.MaxInt; offset += max {
			offsets <- offset
		}
		close(offsets)
	}()

	userInfos := make(chan UserInfo)
	userNames := make(chan OffsetStrings)
	errs := make(chan error)
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				col := makeClientStatsCollector("AllFollowersParallel")
				for offset := range offsets {
					fs, err := c.GetFollowers(username, FollowersOffset(offset), FollowersMax(max))
					if *clientStats {
						col.RecordAndPrint()
					}
					if err != nil {
						errs <- err
						break
					}
					if len(fs) == 0 {
						break
					}
					var us []string
					for _, u := range fs {
						userInfos <- u
						us = append(us, u.Username)
					}
					userNames <- OffsetStrings{Strings: us, Offset: offset}
				}
			}()
		}
		wg.Wait()
		close(userInfos)
		close(userNames)
		close(errs)
	}()

	return userInfos, userNames, errs
}

func (c *Client) AllFollowers(username string, f func(offset int, userInfos UserInfos) error, fOpts ...AllFollowersOption) error {
	opts := MakeAllFollowersOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	for offset := start; ; offset += max {
		followings, err := c.GetFollowers(username, FollowersOffset(offset), FollowersMax(max))
		if err != nil {
			return err
		}
		if len(followings) == 0 {
			break
		}
		if err := f(offset, followings); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) AllFollowingParallel(username string, fOpts ...AllFollowingsOption) (chan UserInfo, chan OffsetStrings, chan error) {
	opts := MakeAllFollowingsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	threads := or.Int(opts.Threads(), defaultThreads)

	userInfos := make(chan UserInfo)
	userNames := make(chan OffsetStrings)
	offsets := make(chan int)
	errs := make(chan error)

	go func() {
		for offset := start; offset < math.MaxInt; offset += max {
			offsets <- offset
		}
		close(offsets)
	}()

	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				col := makeClientStatsCollector("AllFollowingParallel")
				for offset := range offsets {
					fs, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
					if *clientStats {
						col.RecordAndPrint()
					}
					if err != nil {
						errs <- err
						break
					}
					if len(fs) == 0 {
						break
					}
					var us []string
					for _, u := range fs {
						userInfos <- u
						us = append(us, u.Username)
					}
					userNames <- OffsetStrings{Strings: us, Offset: offset}
				}
			}()
		}
		wg.Wait()
		close(userInfos)
		close(userNames)
		close(errs)
	}()

	return userInfos, userNames, errs
}

type CreatePostInfo struct {
	CDate IntDate `json:"cdate"`
	UDate IntDate `json:"udate"`
	UID   string  `json:"uid"`
	Type  string  `json:"_t"`
	ID    string  `json:"_id"`
	Text  string  `json:"txt"`
}

func (c *Client) CreatePost(text string) (CreatePostInfo, error) {
	date := int(time.Now().UnixMilli())
	content := fmt.Sprintf(
		`{"data":{"acl":{"_t":"acl"},"_t":"post","txt":"%s","udate":%d,"cdate":%d,"uid":"%s"},"aux":null,"serial":"post"}`,
		text, date, date, c.username)
	data := url.Values{}
	data.Set("content", content)
	route := "u/post"
	var payload struct {
		Data CreatePostInfo `json:"data"`
	}
	if err := c.post(route, &payload, strings.NewReader(data.Encode())); err != nil {
		return CreatePostInfo{}, err
	}
	return payload.Data, nil
}

type DeletePostInfo struct {
	CDate IntDate `json:"cdate"`
	UDate IntDate `json:"udate"`
	UID   string  `json:"uid"`
	Type  string  `json:"_t"`
	ID    string  `json:"_id"`
	Text  string  `json:"txt"`
}

func (c *Client) DeletePost(postID string) (bool, error) {
	route := fmt.Sprintf("u/post/%s", postID)
	var payload bool
	if err := c.delete(route, &payload); err != nil {
		return false, err
	}
	return payload, nil
}
