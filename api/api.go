package api

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/log"
	"github.com/spudtrooper/gettr/util"
	"github.com/spudtrooper/goutil/flags"
	"github.com/spudtrooper/goutil/must"
)

var (
	clientVerbose = flags.Bool("client_verbose", "verbose client messages")
	user          = flags.String("user", "auth username")
	token         = flags.String("token", "auth token")
	userCreds     = flag.String("user_creds", ".user_creds.json", "file with user credentials")
	clientDebug   = flags.Bool("client_debug", "whether to debug requests")
	requestStats  = flags.Bool("request_stats", "print verbose debugging of request timing")
)

// type Core represents the core gettr Core
type Core struct {
	username string
	xAppAuth string
	debug    bool
}

func (c *Core) Username() string { return c.username }

func MakeClientFromFlags() (*Core, error) {
	if *user != "" && *token != "" {
		client := MakeClient(*user, *token, MakeClientDebug(*clientDebug))
		return client, nil
	}
	if *userCreds != "" {
		client, err := MakeClientFromFile(*userCreds, MakeClientDebug(*clientDebug))
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return nil, errors.Errorf("Must set --user & --token or --creds_file")
}

func MakeClient(user, token string, mOpts ...MakeClientOption) *Core {
	opts := MakeMakeClientOptions(mOpts...)
	xAppAuth := fmt.Sprintf(`{"user": "%s", "token": "%s"}`, user, token)
	return &Core{
		username: user,
		xAppAuth: xAppAuth,
		debug:    opts.Debug(),
	}
}

func readCreds(credsFile string) (user string, token string, ret error) {
	credsBytes, err := ioutil.ReadFile(credsFile)
	if err != nil {
		ret = err
		return
	}
	var creds struct {
		User  string `json:"user"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(credsBytes, &creds); err != nil {
		ret = err
		return
	}
	user, token = creds.User, creds.Token
	return
}

func MakeClientFromFile(credsFile string, mOpts ...MakeClientOption) (*Core, error) {
	opts := MakeMakeClientOptions(mOpts...)
	user, token, err := readCreds(credsFile)
	if err != nil {
		return nil, err
	}
	xAppAuth := fmt.Sprintf(`{"user": "%s", "token": "%s"}`, user, token)
	return &Core{
		username: user,
		xAppAuth: xAppAuth,
		debug:    opts.Debug(),
	}, nil
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
		s := fmt.Sprintf("%s=%s", p.key, url.QueryEscape(fmt.Sprintf("%v", p.val)))
		ss = append(ss, s)
	}
	return fmt.Sprintf("%s?%s", base, strings.Join(ss, "&"))
}

func (c *Core) get(route string, result interface{}) error {
	return c.request("GET", route, result, nil)
}

func (c *Core) post(route string, result interface{}, body io.Reader) error {
	return c.request("POST", route, result, body)
}

func (c *Core) delete(route string, result interface{}) error {
	return c.request("DELETE", route, result, nil)
}

func (c *Core) request(method, route string, result interface{}, body io.Reader) error {
	url := fmt.Sprintf("https://api.gettr.com/%s", route)
	if *clientVerbose {
		// This is to pull off the offsets for debugging and show them to the right of the URL
		var largeNumbers []string
		re := regexp.MustCompile(`(\d+)`)
		for _, m := range re.FindAllStringSubmatch(url, -1) {
			n := must.Atoi(m[1])
			if n < 1000 {
				continue
			}
			d := util.FormatNumber(n)
			largeNumbers = append(largeNumbers, color.YellowString(d))
		}
		log.Printf("requesting %s %s", url, strings.Join(largeNumbers, " "))
	}

	start := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	req.Header.Set("x-app-auth", c.xAppAuth)
	req.Header.Set("userid", c.username)
	if body != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if err != nil {
		return err
	}
	doRes, err := client.Do(req)
	if err != nil {
		return err
	}
	reqStop := time.Now()

	data, err := ioutil.ReadAll(doRes.Body)
	if err != nil {
		return err
	}

	doRes.Body.Close()

	readStop := time.Now()

	if *requestStats {
		reqDur := reqStop.Sub(start)
		readDur := readStop.Sub(reqStop)
		totalDur := readStop.Sub(start)
		log.Printf("request stats: total:%v request:%v read:%v", totalDur, reqDur, readDur)
	}

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
