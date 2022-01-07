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

func (c *Client) GetUserInfo(username string) (*UserInfo, error) {
	route := fmt.Sprintf("s/uinf/%s", username)
	var payload struct {
		Data UserInfo
	}
	if err := c.request(route, &payload); err != nil {
		return nil, err
	}
	return &payload.Data, nil
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
	if err := c.request(route, &payload); err != nil {
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

type Suggestions struct {
	HTInfo []HtInfo
}

func (c *Client) GetSuggestions(sOpts ...SuggestOption) (*Suggestions, error) {
	opts := MakeSuggestOptions(sOpts...)
	max := or.Int(opts.Max(), 10)
	route := createRoute("s/hashtag/suggest", param{"max", max})
	type suggestions struct {
		HTInfo map[string]HtInfo `json:"htinfo"`
	}
	var payload struct {
		Aux suggestions `json:"aux"`
	}
	if err := c.request(route, &payload); err != nil {
		return nil, err
	}
	var res Suggestions
	for _, s := range payload.Aux.HTInfo {
		res.HTInfo = append(res.HTInfo, s)
	}
	return &res, nil
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

type Posts struct {
	Posts []PostInfo
}

func (c *Client) GetPosts(username string, pOpts ...PostsOption) (*Posts, error) {
	opts := MakePostsOptions(pOpts...)
	offset := or.Int(opts.Offset(), 0)
	max := or.Int(opts.Max(), offset+20)
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
	if err := c.request(route, &payload); err != nil {
		return nil, err
	}
	var res Posts
	for _, p := range payload.Aux.Posts {
		res.Posts = append(res.Posts, p)
	}
	return &res, nil
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

type Comments struct {
	Comments []CommentInfo
}

func (c *Client) GetComments(post string, cOpts ...CommentsOption) (*Comments, error) {
	opts := MakeCommentsOptions(cOpts...)
	offset := or.Int(opts.Offset(), 0)
	max := or.Int(opts.Max(), offset+20)
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
	if err := c.request(route, &payload); err != nil {
		return nil, err
	}
	var res Comments
	for _, p := range payload.Aux.Comments {
		res.Comments = append(res.Comments, p)
	}
	return &res, nil
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
	Update    StringDate `json:"update"`
	Type      string     `json:"_t"`
	ID        string     `json:"_id"`
}

type PostDetails struct {
	ShareInfo
	PostInfo
	UserInfos []UserInfo
}

func (c *Client) GetPost(post string, pOpts ...PostOption) (*PostDetails, error) {
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
	if err := c.request(route, &payload); err != nil {
		return nil, err
	}
	res := PostDetails{
		ShareInfo: payload.Aux.Share,
		PostInfo:  payload.Data,
	}
	for _, u := range payload.Aux.Uinf {
		res.UserInfos = append(res.UserInfos, u)
	}
	return &res, nil
}
