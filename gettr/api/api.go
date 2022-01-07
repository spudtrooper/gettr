package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spudtrooper/goutil/or"
)

type Client struct {
	xAppAuth string
	debug    bool
}

func MakeClient(user, token string, mOpts ...MakeClientOption) *Client {
	opts := MakeMakeClientOptions(mOpts...)
	xAppAuth := fmt.Sprintf(`{"user": "%s", "token": "%s"}`, user, token)
	return &Client{
		xAppAuth: xAppAuth,
		debug:    opts.Debug(),
	}
}

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

func (c *Client) request(route string, result interface{}) error {
	url := fmt.Sprintf("https://api.gettr.com/%s", route)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("x-app-auth", c.xAppAuth)
	if err != nil {
		return err
	}
	doRes, err := client.Do(req)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(doRes.Body)
	if err != nil {
		return err
	}

	doRes.Body.Close()

	if c.debug {
		prettyJSON, err := prettyPrintJSON(data)
		if err != nil {
			return err
		}
		log.Printf("from route %q have response %s", route, prettyJSON)
	}

	var payload struct {
		ResponseCode string `json:"rc"`
		Result       interface{}
	}
	payload.Result = result
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if payload.ResponseCode != "OK" {
		return errors.Errorf("non-OK response: %+v", payload)
	}

	return nil
}

func prettyPrintJSON(b []byte) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, b, "", "\t"); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
