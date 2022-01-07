package api

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spudtrooper/goutil/or"
)

type UserInfo struct {
	Nickname string
	Username string
	Lang     string
	Status   string
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
	route := fmt.Sprintf("s/hashtag/suggest?max=%d", max)
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
