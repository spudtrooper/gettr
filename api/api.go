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
	"github.com/spudtrooper/goutil/or"
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
	username  string
	xAppAuth  string
	debug     bool
	authToken string
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
		username:  user,
		xAppAuth:  xAppAuth,
		debug:     opts.Debug(),
		authToken: token,
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
		username:  user,
		xAppAuth:  xAppAuth,
		debug:     opts.Debug(),
		authToken: token,
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

func (c *Core) get(route string, result interface{}, rOpts ...RequestOption) (*http.Response, error) {
	return c.request("GET", route, result, nil, rOpts...)
}

// TODO: Move body to a RequestOption
func (c *Core) post(route string, result interface{}, body io.Reader, rOpts ...RequestOption) (*http.Response, error) {
	return c.request("POST", route, result, body, rOpts...)
}

// TODO: Move body to a RequestOption
func (c *Core) patch(route string, result interface{}, body io.Reader, rOpts ...RequestOption) (*http.Response, error) {
	return c.request("PATCH", route, result, body, rOpts...)
}

func (c *Core) delete(route string, result interface{}, rOpts ...RequestOption) (*http.Response, error) {
	return c.request("DELETE", route, result, nil, rOpts...)
}

func (c *Core) request(method, route string, result interface{}, body io.Reader, rOpts ...RequestOption) (*http.Response, error) {
	opts := MakeRequestOptions(rOpts...)
	host := or.String(opts.Host(), "api.gettr.com")
	url := fmt.Sprintf("https://%s/%s", host, route)
	if *clientVerbose {
		// This is to pull off the offsets for debugging and show them to the right of the URL
		var largeNumbers []string
		if strings.Contains(url, "offset") {
			re := regexp.MustCompile(`(\d+)`)
			for _, m := range re.FindAllStringSubmatch(url, -1) {
				n := must.Atoi(m[1])
				if n < 1000 {
					continue
				}
				d := util.FormatNumber(n)
				largeNumbers = append(largeNumbers, color.YellowString(d))
			}
		}
		log.Printf("requesting %s %s", url, strings.Join(largeNumbers, " "))
		if len(opts.ExtraHeaders()) > 0 {
			log.Printf("  with extra headers:")
			for k, v := range opts.ExtraHeaders() {
				log.Printf("    %s: %s", k, v)
			}
		}
	}

	start := time.Now()

	client := &http.Client{}
	if opts.NoRedirect() {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-app-auth", c.xAppAuth)
	for k, v := range opts.ExtraHeaders() {
		req.Header.Set(k, v)
	}
	if c.debug {
		log.Printf("requesting %s", url)
		log.Printf("  headers:")
		for k, v := range req.Header {
			log.Printf("    %s: %s", k, v)
		}
		log.Printf("  body: %v", body)
	}
	doRes, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	reqStop := time.Now()

	data, err := ioutil.ReadAll(doRes.Body)
	if err != nil {
		return nil, err
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
		log.Printf("response <<<\n%s\n>>>", string(data))
	}

	if c := string(data); strings.Contains(c, "Request unsuccessful") {
		return nil, errors.Errorf("LIMITED: Request unsuccessful. Incapsula incident")
	}
	if c := string(data); strings.Contains(c, "<HTML><HEAD><TITLE>Loading</TITLE>") {
		return nil, errors.Errorf("LIMITED: Loading instead")
	}

	if c.debug {
		prettyJSON, err := prettyPrintJSON(data)
		if err != nil {
			log.Printf("ignoring prettyPrintJSON error: %v", err)
			prettyJSON = string(data)
		}
		log.Printf("from route %q have response %s", route, prettyJSON)
	}

	if len(data) > 0 {
		if opts.CustomPayload() != nil {
			if err := json.Unmarshal(data, opts.CustomPayload()); err != nil {
				return nil, err
			}
		} else {
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
				return nil, err
			}
			if *clientVerbose {
				log.Printf("got response with rc=%s", payload.ResponseCode)
			}
			if payload.ResponseCode != "OK" {
				return nil, errors.Errorf("response error: %+v", payload.Error)
			}
		}
	}

	return doRes, nil
}

func prettyPrintJSON(b []byte) (string, error) {
	b = []byte(strings.TrimSpace(string(b)))
	if len(b) == 0 {
		return "", nil
	}
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, b, "", "\t"); err != nil {
		return "", errors.Errorf("json.Indent: payload=%q: %v", string(b), err)
	}
	return prettyJSON.String(), nil
}

func userURI(username string) string {
	return fmt.Sprintf("https://gettr.com/user/%s", username)
}

func postURI(postID string) string {
	return fmt.Sprintf("https://gettr.com/post/%s", postID)
}

// https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and/28596225
func jsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(t); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
