package model

import (
	"reflect"
	"sort"
	"testing"

	"github.com/spudtrooper/goutil/io"
)

func TestGetAllStrings(t *testing.T) {
	var dir string
	// Set this to true to persist the directory, in case you want to debug this test.
	if false {
		dir = "model/tempDir"
		if _, err := io.MkdirAll(dir); err != nil {
			t.Fatalf("MkdirAll(%q): %v", dir, err)
		}
	} else {
		dir = t.TempDir()
	}
	c := makeCache(dir)

	if err := c.SetGeneric([]string{"1", "2"}, "users", "foo", "dir", "1"); err != nil {
		t.Fatalf("SetGeneric: %v", err)
	}
	if err := c.SetGeneric([]string{"3", "4"}, "users", "foo", "dir", "2"); err != nil {
		t.Fatalf("SetGeneric: %v", err)
	}
	if err := c.SetGeneric([]string{"5", "6"}, "users", "foo", "dir", "3"); err != nil {
		t.Fatalf("SetGeneric: %v", err)
	}

	{
		strings, err := c.GetAllStrings("users", "foo", "dir")
		if err != nil {
			t.Fatalf("GetAllStrings: %v", err)
		}
		sort.Strings(strings)
		if got, want := strings, []string{"1", "2", "3", "4", "5", "6"}; !reflect.DeepEqual(want, got) {
			t.Fatalf("GetAllStrings: want != got: %v %v", want, got)
		}
	}
	{
		strings, err := c.FindKeys("users", "foo", "dir")
		if err != nil {
			t.Fatalf("FindKeys: %v", err)
		}
		sort.Strings(strings)
		if got, want := strings, []string{"1", "2", "3"}; !reflect.DeepEqual(want, got) {
			t.Fatalf("FindKeys: want != got: %v %v", want, got)
		}
	}
}
