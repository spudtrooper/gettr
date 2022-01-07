package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
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
