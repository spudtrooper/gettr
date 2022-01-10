package api

import (
	"testing"
)

func TestGetUserInfo(t *testing.T) {
	c, err := MakeClientFromFile("../.user_creds.json", MakeClientDebug(true))
	if err != nil {
		t.Fatalf("cannot make client: %v", err)
	}

	info, err := c.GetUserInfo("spudtrooper")
	if err != nil {
		t.Fatalf("cannot request GetUserInfo: %v", err)
	}
	if got, want := info.Username, "spudtrooper"; got != want {
		t.Errorf("got %q but expected %q", got, want)
	}
}
