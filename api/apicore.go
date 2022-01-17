package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
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

func (c *Core) GetUserInfo(username string) (UserInfo, error) {
	route := fmt.Sprintf("s/uinf/%s", username)
	var payload struct {
		Data UserInfo
	}
	if _, err := c.get(route, &payload); err != nil {
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

func (c *Core) GetPublicGlobals() (*PublicGlobals, error) {
	route := "u/public_globals"
	var payload struct {
		Globals PublicGlobals
	}
	if _, err := c.get(route, &payload); err != nil {
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

func (c *Core) GetSuggestions(sOpts ...SuggestOption) ([]HtInfo, error) {
	opts := MakeSuggestOptions(sOpts...)
	max := or.Int(opts.Max(), 10)
	route := createRoute("s/hashtag/suggest", param{"max", max})
	type suggestions struct {
		HTInfo map[string]HtInfo `json:"htinfo"`
	}
	var payload struct {
		Aux suggestions `json:"aux"`
	}
	if _, err := c.get(route, &payload); err != nil {
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

func (p PostInfo) URI() string { return postURI(p.ID) }

func (c *Core) getPosts(route string, rOpts ...RequestOption) ([]PostInfo, error) {
	type posts struct {
		Posts map[string]PostInfo `json:"post"`
	}
	var payload struct {
		Aux posts `json:"aux"`
	}
	if _, err := c.get(route, &payload, rOpts...); err != nil {
		return nil, err
	}
	var res []PostInfo
	for _, p := range payload.Aux.Posts {
		res = append(res, p)
	}
	return res, nil
}

func (c *Core) GetPosts(username string, pOpts ...PostsOption) ([]PostInfo, error) {
	opts := MakePostsOptions(pOpts...)
	offset := or.Int(opts.Offset(), defaultOffset)
	max := or.Int(opts.Max(), defaultMax)
	dir := or.String(opts.Dir(), defaultDir)
	incl := or.String(strings.Join(opts.Incl(), "|"), "posts|stats|userinfo|shared|liked")
	fp := or.String(opts.Fp(), "f_uo")
	route := createRoute(fmt.Sprintf("u/user/%s/posts", username),
		param{"offset", offset}, param{"max", max}, param{"dir", dir}, param{"incl", incl}, param{"fp", fp})
	return c.getPosts(route)
}

func (c *Core) Timeline(pOpts ...TimelineOption) ([]PostInfo, error) {
	opts := MakeTimelineOptions(pOpts...)
	offset := or.Int(opts.Offset(), defaultOffset)
	max := or.Int(opts.Max(), defaultMax)
	dir := or.String(opts.Dir(), defaultDir)
	incl := or.String(strings.Join(opts.Incl(), "|"), "posts|stats|userinfo|shared|liked")
	merge := or.String(opts.Merge(), "shares")
	route := createRoute(fmt.Sprintf("u/user/%s/timeline", c.username),
		param{"offset", offset}, param{"max", max}, param{"dir", dir}, param{"incl", incl}, param{"merge", merge})
	return c.getPosts(route)
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

func (c *Core) GetComments(post string, cOpts ...CommentsOption) ([]CommentInfo, error) {
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
	if _, err := c.get(route, &payload); err != nil {
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
func (u UserInfo) URI() string           { return userURI(u.ID) }

type UserInfos []UserInfo

type PostDetails struct {
	ShareInfo
	PostInfo
	UserInfos
}

func (c *Core) GetPost(post string, pOpts ...PostOption) (PostDetails, error) {
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
	if _, err := c.get(route, &payload); err != nil {
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

func (c *Core) GetMuted(mOpts ...MutedOption) (UserInfos, error) {
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
	if _, err := c.get(route, &payload); err != nil {
		return nil, err
	}
	var res UserInfos
	for _, u := range payload.Aux.Uinf {
		res = append(res, u)
	}
	return res, nil
}

func (c *Core) GetFollowings(username string, fOpts ...FollowingsOption) (UserInfos, error) {
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
	if _, err := c.get(route, &payload); err != nil {
		return nil, err
	}
	var res UserInfos
	for _, u := range payload.Aux.Uinf {
		res = append(res, u)
	}
	return res, nil
}

func (c *Core) Follow(username string) error {
	route := createRoute(fmt.Sprintf("u/user/%s/follows/%s", c.username, username))
	if _, err := c.post(route, nil, nil); err != nil {
		return err
	}
	return nil
}

func (c *Core) GetFollowers(username string, fOpts ...FollowersOption) (UserInfos, error) {
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
	if _, err := c.get(route, &payload); err != nil {
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

type CreatePostInfo struct {
	CDate IntDate `json:"cdate"`
	UDate IntDate `json:"udate"`
	UID   string  `json:"uid"`
	Type  string  `json:"_t"`
	ID    string  `json:"_id"`
	Text  string  `json:"txt"`
}

func (c CreatePostInfo) URI() string { return postURI(c.ID) }

func (c *Core) CreatePost(text string, cOpts ...CreatePostOption) (CreatePostInfo, error) {
	opts := MakeCreatePostOptions(cOpts...)
	date := int(time.Now().UnixMilli())
	var contentData struct {
		Data struct {
			ACL struct {
				Type string `json:"_t"`
			} `json:"acl"`
			Type          string   `json:"_t"`
			Text          string   `json:"txt"`
			Description   string   `json:"dsc"`
			UDate         IntDate  `json:"udate"`
			CDate         IntDate  `json:"cdate"`
			UID           string   `json:"uid"`
			Images        []string `json:"imgs"`
			PreviewImage  string   `json:"previmg"`
			PreviewSource string   `json:"prevsrc"`
			VidWidth      int      `json:"vid_wid"`
			VidHeight     int      `json:"vid_hgt"`
			Title         string   `json:"ttl"`
		} `json:"data"`
		Aux    interface{} `json:"aux"`
		Serial string      `json:"serial"`
	}
	contentData.Data.ACL.Type = "acl"
	contentData.Data.Type = "post"
	contentData.Data.Text = text
	contentData.Data.CDate = IntDate(date)
	contentData.Data.UDate = IntDate(date)
	contentData.Data.UID = c.username
	contentData.Serial = "post"
	if len(opts.Images()) > 0 {
		contentData.Data.Images = opts.Images()
		contentData.Data.VidWidth = 152
		contentData.Data.VidHeight = 250
	}
	contentData.Data.VidWidth = 152
	contentData.Data.VidHeight = 250
	contentData.Data.Description = opts.Description()
	contentData.Data.PreviewImage = opts.PreviewImage()
	contentData.Data.PreviewSource = opts.PreviewSource()
	contentData.Data.Title = opts.Title()
	contentBytes, err := json.Marshal(&contentData)
	if err != nil {
		return CreatePostInfo{}, err
	}
	content := string(contentBytes)
	if opts.Debug() {
		log.Printf("DEBUG: CreatePost content: <<<\n\n%s\n\n>>>", content)
	}
	data := url.Values{}
	data.Set("content", content)
	route := "u/post"
	var payload struct {
		Data CreatePostInfo `json:"data"`
	}
	extraHeaders := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	if _, err := c.post(route, &payload, strings.NewReader(data.Encode()), RequestExtraHeaders(extraHeaders)); err != nil {
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

func (c *Core) DeletePost(postID string) (bool, error) {
	route := fmt.Sprintf("u/post/%s", postID)
	var payload bool
	if _, err := c.delete(route, &payload); err != nil {
		return false, err
	}
	return payload, nil
}

func base64String(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

type UploadInfo struct {
	ORI     string `json:"ori"`
	MD5     string `json:"md5"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (c *Core) Upload(f string) (UploadInfo, error) {
	extWithNoDot := string(path.Ext(f)[1:])
	filename := "53d4e55-65f-07ee-ea2f-e3cc0aeec16-base64image." + extWithNoDot
	filetype := "image/" + extWithNoDot
	token := c.authToken
	body, err := ioutil.ReadFile(f)
	if err != nil {
		return UploadInfo{}, err
	}
	var patchLocation string
	commonHeaders := func(m map[string]string) map[string]string {
		common := map[string]string{
			"userid":        c.username,
			"filename":      filename,
			"Authorization": token,
			"Tus-Resumable": `1.0.0`,
		}
		for k, v := range common {
			m[k] = v
		}
		return m
	}
	{
		uploadMetadata := fmt.Sprintf("filename %s,filetype %s", base64String(filename), base64String(filetype))
		extraHeaders := commonHeaders(map[string]string{
			"Upload-Metadata": uploadMetadata,
			"Content-Length":  `0`,
			"Upload-Length":   fmt.Sprintf("%d", len(body)),
		})
		route := "media/big/upload"
		res, err := c.post(route, nil, nil, RequestExtraHeaders(extraHeaders), RequestHost("upload.gettr.com"))
		if err != nil {
			return UploadInfo{}, err
		}
		loc := res.Header.Get("Location")
		if loc == "" {
			return UploadInfo{}, errors.Errorf("no location from the POST: response=%v", res)
		}
		patchLocation = loc
	}
	{
		extraHeaders := commonHeaders(map[string]string{
			"Content-Type":   `application/offset+octet-stream`,
			"Content-Length": fmt.Sprintf("%d", len(body)),
			"Upload-Offset":  `0`,
		})
		route := patchLocation
		if strings.HasPrefix(route, "/") {
			route = string(route[1:])
		}
		res := UploadInfo{}
		if _, err := c.patch(route, nil, bytes.NewBuffer(body), RequestExtraHeaders(extraHeaders), RequestHost("upload.gettr.com"), RequestCustomPayload(&res)); err != nil {
			return UploadInfo{}, err
		}
		return res, nil
	}
}

type UpdateProfileInfo struct{ UserInfo }

func (c *Core) UpdateProfile(pOpts ...UpdateProfileOption) (UpdateProfileInfo, error) {
	opts := MakeUpdateProfileOptions(pOpts...)
	data := url.Values{}
	if opts.Description() != "" {
		data.Set("dsc", opts.Description())
	}
	if opts.BackgroundImage() != "" {
		data.Set("bgimg", opts.BackgroundImage())
	}
	if opts.Icon() != "" {
		data.Set("ico", opts.Icon())
	}
	if opts.Website() != "" {
		data.Set("website", opts.Website())
	}
	if opts.Location() != "" {
		data.Set("location", opts.Location())
	}
	route := fmt.Sprintf("u/user/%s/profile", c.username)
	var payload struct {
		Data UpdateProfileInfo `json:"data"`
	}
	extraHeaders := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	if _, err := c.post(route, &payload, strings.NewReader(data.Encode()), RequestExtraHeaders(extraHeaders)); err != nil {
		return UpdateProfileInfo{}, err
	}
	return payload.Data, nil
}

func (c *Core) LikePost(postID string) error {
	route := fmt.Sprintf("u/user/%s/likes/post/%s", c.username, postID)
	if _, err := c.post(route, nil, nil); err != nil {
		return err
	}
	return nil
}

func (c *Core) makeSearchBody(query string, sOpts ...SearchOption) ([]byte, error) {
	opts := MakeSearchOptions(sOpts...)
	incl := or.String(strings.Join(opts.Incl(), "|"), "poststats|shared|liked|posts|userinfo|followings|followers|videos")
	max := or.Int(opts.Max(), defaultMax)
	offset := or.Int(opts.Offset(), defaultOffset)
	type content struct {
		Query  string `json:"q"`
		Incl   string `json:"incl"`
		Max    int    `json:"max"`
		Offset int    `json:"offset"`
	}
	cd := struct {
		Content content `json:"content"`
	}{
		Content: content{
			Query:  query,
			Incl:   incl,
			Max:    max,
			Offset: offset,
		},
	}
	body, err := json.Marshal(&cd)
	if err != nil {
		return nil, err
	}
	if opts.Debug() {
		log.Printf("DEBUG: Search body: <<<\n\n%s\n\n>>>", string(body))
	}
	return body, err
}

func (c *Core) SearchPosts(query string, sOpts ...SearchOption) ([]PostInfo, error) {
	body, err := c.makeSearchBody(query, sOpts...)
	if err != nil {
		return nil, err
	}
	route := "u/posts/srch/phrase"
	extraHeaders := map[string]string{
		"content-type": `application/json`,
	}
	type posts struct {
		Posts map[string]PostInfo `json:"post"`
	}
	var payload struct {
		Aux posts `json:"aux"`
	}
	if _, err := c.post(route, &payload, bytes.NewBuffer(body), RequestExtraHeaders(extraHeaders)); err != nil {
		return nil, err
	}
	var res []PostInfo
	for _, p := range payload.Aux.Posts {
		res = append(res, p)
	}
	return res, nil
}

func (c *Core) SearchUsers(query string, sOpts ...SearchOption) ([]UserInfo, error) {
	body, err := c.makeSearchBody(query, sOpts...)
	if err != nil {
		return nil, err
	}
	route := "u/users/srch/phrase"
	extraHeaders := map[string]string{
		"content-type": `application/json`,
	}
	type users struct {
		Uinf map[string]UserInfo `json:"uinf"`
	}
	var payload struct {
		Aux users `json:"aux"`
	}
	if _, err := c.post(route, &payload, bytes.NewBuffer(body), RequestExtraHeaders(extraHeaders)); err != nil {
		return nil, err
	}
	var res []UserInfo
	for _, p := range payload.Aux.Uinf {
		res = append(res, p)
	}
	return res, nil
}
