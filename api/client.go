package api

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

var (
	clientVerbose = flag.Bool("client_verbose", false, "verbose client messages")
)

type Client struct {
	username string
	xAppAuth string
	debug    bool
}

func MakeClient(user, token string, mOpts ...MakeClientOption) *Client {
	opts := MakeMakeClientOptions(mOpts...)
	xAppAuth := fmt.Sprintf(`{"user": "%s", "token": "%s"}`, user, token)
	return &Client{
		username: user,
		xAppAuth: xAppAuth,
		debug:    opts.Debug(),
	}
}

type param struct {
	key string
	val interface{}
}

func createRoute(base string, ps ...param) string {
	if len(ps) == 0 {
		return base
	}
	var ss []string
	for _, p := range ps {
		s := fmt.Sprintf("%s=%v", p.key, p.val)
		ss = append(ss, s)
	}
	return fmt.Sprintf("%s?%s", base, strings.Join(ss, "&"))
}

func (c *Client) get(route string, result interface{}) error {
	return c.request("GET", route, result)
}

func (c *Client) post(route string, result interface{}) error {
	return c.request("POST", route, result)
}

func (c *Client) request(method, route string, result interface{}) error {
	url := fmt.Sprintf("https://api.gettr.com/%s", route)
	if *clientVerbose {
		log.Printf("requesting %s", url)
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
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
		Error        struct {
			Code string `json:"code"`
			EMsg string `json:"emsg"`
			Type string `json:"_t"`
		} `json:"error"`
		Result interface{}
	}
	payload.Result = result
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if payload.ResponseCode != "OK" {
		return errors.Errorf("response error: %+v", payload.Error)
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
