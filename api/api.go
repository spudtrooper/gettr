package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/or"
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
	offset := or.Int(opts.Offset(), 0)
	max := or.Int(opts.Max(), 20)
	dir := or.String(opts.Dir(), "fwd")
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
	offset := or.Int(opts.Offset(), 0)
	max := or.Int(opts.Max(), 20)
	dir := or.String(opts.Dir(), "fwd")
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
	TwtFlg    string     `json:"twt_flg"`
	TwtFlw    string     `json:"twt_flw"`
	Flg       string     `json:"flg"`
	Flw       string     `json:"flw"`
	CDate     StringDate `json:"cdate"`
	UDate     StringDate `json:"udate"`
	Type      string     `json:"_t"`
	ID        string     `json:"_id"`
}

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
	offset := or.Int(opts.Offset(), 0)
	max := or.Int(opts.Max(), 20)
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
	offset := or.Int(opts.Offset(), 0)
	max := or.Int(opts.Max(), 20)
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
	var res UserInfos
	max := 20
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

func (c *Client) AllFollowings(username string, f func(UserInfos) error, fOpts ...FollowingsOption) error {
	max := 20
	for offset := 0; ; offset += max {
		followings, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
		if err != nil {
			return err
		}
		if len(followings) == 0 {
			break
		}
		if err := f(followings); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Follow(username string) error {
	route := createRoute(fmt.Sprintf("u/user/%s/follows/%s", c.username, username))
	if err := c.post(route, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetFollowers(username string, fOpts ...FollowersOption) (UserInfos, error) {
	opts := MakeFollowersOptions(fOpts...)
	offset := or.Int(opts.Offset(), 0)
	max := or.Int(opts.Max(), 20)
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

func (c *Client) AllFollowers(username string, f func(UserInfos) error, fOpts ...FollowersOption) error {
	max := 20
	for offset := 0; ; offset += max {
		followings, err := c.GetFollowers(username, FollowersOffset(offset), FollowersMax(max))
		if err != nil {
			return err
		}
		if len(followings) == 0 {
			break
		}
		if err := f(followings); err != nil {
			return err
		}
	}
	return nil
}
