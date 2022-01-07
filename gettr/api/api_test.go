package api

import (
	"os"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func TestGetUserInfo(t *testing.T) {
	c, err := makeClient()
	if err != nil {
		t.Fatalf("cannot make client: %v", err)
	}

	info, err := c.GetUserInfo("spudtrooper")
	if err != nil {
		t.Fatalf("cannot request GetUserInfo: %v", err)
	}
	want := UserInfo{
		Nickname: "spudtrooper",
		Username: "spudtrooper",
		Lang:     "en_us",
		Status:   "a",
	}
	if got := *info; !reflect.DeepEqual(want, got) {
		t.Errorf("got %v  but expected %v", got, want)
	}

}

func makeClient() (*Client, error) {
	user := "spudtrooper"
	token, ok := os.LookupEnv("GETTR_TOKEN")
	if !ok {
		return nil, errors.Errorf("set GETTR_TOKEN to your gettr auth token")
	}
	c := MakeClient(user, token, MakeClientDebug(true))
	return c, nil
}
