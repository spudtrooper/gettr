package model

import (
	"context"
	"reflect"
	"testing"

	"github.com/spudtrooper/gettr/api"
)

const (
	testUsername = "_____testUsername______"
)

func TestDBUserInfo(t *testing.T) {
	*dbVerboseUserInfo = true

	ctx := context.Background()

	db, err := MakeDB(ctx, MakeDBDbName("gettrtest"))
	if err != nil {
		t.Fatalf("MakeDB: %v", err)
	}
	defer db.Disconnect(ctx)

	if err := db.deleteAllUserInfo(ctx); err != nil {
		t.Fatalf("deleteAllUserInfo: %v", err)
	}

	username := testUsername

	userInfo := api.UserInfo{
		Username: username,
		Lang:     "en",
	}
	if err := db.SetUserInfo(ctx, username, userInfo); err != nil {
		t.Fatalf("SetUserInfo: %v", err)
	}

	got, err := db.GetUserInfo(ctx, username)
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}

	if got, want := *got, userInfo; !reflect.DeepEqual(want, got) {
		t.Errorf("GetUserInfo: want != got: %v %v", want, got)
	}
}

func TestDBFollowers(t *testing.T) {
	*dbVerboseFollowers = true

	ctx := context.Background()

	db, err := MakeDB(ctx, MakeDBDbName("gettrtest"))
	if err != nil {
		t.Fatalf("MakeDB: %v", err)
	}
	defer db.Disconnect(ctx)

	if err := db.deleteAllFollowers(ctx); err != nil {
		t.Fatalf("deleteAllFollowers: %v", err)
	}

	username := testUsername

	followers1 := []string{"11", "12", "13"}
	followers2 := []string{"21", "22", "23"}
	followers3 := []string{"31", "32", "33"}

	if err := db.SetFollowers(ctx, username, 1, followers1); err != nil {
		t.Fatalf("SetFollowers: %v", err)
	}
	if err := db.SetFollowers(ctx, username, 2, followers2); err != nil {
		t.Fatalf("SetFollowers: %v", err)
	}
	if err := db.SetFollowers(ctx, username, 3, followers3); err != nil {
		t.Fatalf("SetFollowers: %v", err)
	}

	var want []string
	want = append(want, followers1...)
	want = append(want, followers2...)
	want = append(want, followers3...)

	{
		got, err := db.GetFollowersSync(ctx, username)
		if err != nil {
			t.Fatalf("GetFollowersSync: %v", err)
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("GetFollowersSync: want != got: %v %v", want, got)
		}
	}

	{
		followers, errors, err := db.GetFollowers(ctx, username)
		if err != nil {
			t.Fatalf("GetFollowers: %v", err)
		}
		var got []string
		for f := range followers {
			got = append(got, f)
		}
		for err := range errors {
			t.Fatalf("GetFollowers: %v", err)
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("GetFollowers: want != got: %v %v", want, got)
		}
	}

}

func TestDBFollowing(t *testing.T) {
	*dbVerboseFollowing = true

	ctx := context.Background()

	db, err := MakeDB(ctx, MakeDBDbName("gettrtest"))
	if err != nil {
		t.Fatalf("MakeDB: %v", err)
	}
	defer db.Disconnect(ctx)

	if err := db.deleteAllFollowing(ctx); err != nil {
		t.Fatalf("deleteAllFollowing: %v", err)
	}

	username := testUsername

	following1 := []string{"11", "12", "13"}
	following2 := []string{"21", "22", "23"}
	following3 := []string{"31", "32", "33"}

	if err := db.SetFollowing(ctx, username, 1, following1); err != nil {
		t.Fatalf("SetFollowing: %v", err)
	}
	if err := db.SetFollowing(ctx, username, 2, following2); err != nil {
		t.Fatalf("SetFollowing: %v", err)
	}
	if err := db.SetFollowing(ctx, username, 3, following3); err != nil {
		t.Fatalf("SetFollowing: %v", err)
	}

	var want []string
	want = append(want, following1...)
	want = append(want, following2...)
	want = append(want, following3...)

	following, errors, err := db.GetFollowing(ctx, username)
	if err != nil {
		t.Fatalf("GetFollowing: %v", err)
	}
	var got []string
	for f := range following {
		got = append(got, f)
	}
	for err := range errors {
		t.Fatalf("GetFollowing: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("GetFollowing: want != got: %v %v", want, got)
	}

}

func TestDBSkip(t *testing.T) {
	*dbVerboseUserInfo = true

	ctx := context.Background()

	db, err := MakeDB(ctx, MakeDBDbName("gettrtest"))
	if err != nil {
		t.Fatalf("MakeDB: %v", err)
	}
	defer db.Disconnect(ctx)

	if err := db.deleteAllUserInfo(ctx); err != nil {
		t.Fatalf("deleteAllUserInfo: %v", err)
	}

	username := testUsername

	if got, err := db.GetUserSkip(ctx, username); err != nil {
		t.Fatalf("GetUserSkip: %v", err)
	} else {
		if want := false; got != want {
			t.Errorf("GetUserSkip: got != want: %v != %v", got, want)
		}
	}

	if err := db.SetUserSkip(ctx, username, false); err != nil {
		t.Fatalf("SetUserSkip: %v", err)
	}

	if got, err := db.GetUserSkip(ctx, username); err != nil {
		t.Fatalf("GetUserSkip: %v", err)
	} else {
		if want := false; got != want {
			t.Errorf("GetUserSkip: got != want: %v != %v", got, want)
		}
	}

	if err := db.SetUserSkip(ctx, username, true); err != nil {
		t.Fatalf("SetUserSkip: %v", err)
	}

	if got, err := db.GetUserSkip(ctx, username); err != nil {
		t.Fatalf("GetUserSkip: %v", err)
	} else {
		if want := true; got != want {
			t.Errorf("GetUserSkip: got != want: %v != %v", got, want)
		}
	}
}

func TestDBFollowersDone(t *testing.T) {
	*dbVerboseUserInfo = true

	ctx := context.Background()

	db, err := MakeDB(ctx, MakeDBDbName("gettrtest"))
	if err != nil {
		t.Fatalf("MakeDB: %v", err)
	}
	defer db.Disconnect(ctx)

	if err := db.deleteAllUserInfo(ctx); err != nil {
		t.Fatalf("deleteAllUserInfo: %v", err)
	}

	username := testUsername

	if got, err := db.GetUserFollowersDone(ctx, username); err != nil {
		t.Fatalf("GetUserFollowersDone: %v", err)
	} else {
		if want := false; got != want {
			t.Errorf("GetUserFollowersDone: got != want: %v != %v", got, want)
		}
	}

	if err := db.SetUserFollowersDone(ctx, username, false); err != nil {
		t.Fatalf("SetUserFollowersDone: %v", err)
	}

	if got, err := db.GetUserFollowersDone(ctx, username); err != nil {
		t.Fatalf("GetUserFollowersDone: %v", err)
	} else {
		if want := false; got != want {
			t.Errorf("GetUserFollowersDone: got != want: %v != %v", got, want)
		}
	}

	if err := db.SetUserFollowersDone(ctx, username, true); err != nil {
		t.Fatalf("SetUserFollowersDone: %v", err)
	}

	if got, err := db.GetUserFollowersDone(ctx, username); err != nil {
		t.Fatalf("GetUserFollowersDone: %v", err)
	} else {
		if want := true; got != want {
			t.Errorf("GetUserFollowersDone: got != want: %v != %v", got, want)
		}
	}
}

func TestDBFollowingDone(t *testing.T) {
	*dbVerboseUserInfo = true

	ctx := context.Background()

	db, err := MakeDB(ctx, MakeDBDbName("gettrtest"))
	if err != nil {
		t.Fatalf("MakeDB: %v", err)
	}
	defer db.Disconnect(ctx)

	if err := db.deleteAllUserInfo(ctx); err != nil {
		t.Fatalf("deleteAllUserInfo: %v", err)
	}

	username := testUsername

	if got, err := db.GetUserFollowingDone(ctx, username); err != nil {
		t.Fatalf("GetUserFollowingDone: %v", err)
	} else {
		if want := false; got != want {
			t.Errorf("GetUserFollowingDone: got != want: %v != %v", got, want)
		}
	}

	if err := db.SetUserFollowingDone(ctx, username, false); err != nil {
		t.Fatalf("SetUserFollowingDone: %v", err)
	}

	if got, err := db.GetUserFollowingDone(ctx, username); err != nil {
		t.Fatalf("GetUserFollowingDone: %v", err)
	} else {
		if want := false; got != want {
			t.Errorf("GetUserFollowingDone: got != want: %v != %v", got, want)
		}
	}

	if err := db.SetUserFollowingDone(ctx, username, true); err != nil {
		t.Fatalf("SetUserFollowingDone: %v", err)
	}

	if got, err := db.GetUserFollowingDone(ctx, username); err != nil {
		t.Fatalf("GetUserFollowingDone: %v", err)
	} else {
		if want := true; got != want {
			t.Errorf("GetUserFollowingDone: got != want: %v != %v", got, want)
		}
	}
}
