package model

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/log"
	"github.com/spudtrooper/goutil/flags"
	"github.com/spudtrooper/goutil/or"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	dbName string
	client *mongo.Client
}

func MakeDB(ctx context.Context, mOpts ...MakeDBOption) (*DB, error) {
	opts := MakeMakeDBOptions(mOpts...)

	port := or.Int(opts.Port(), 27017)
	dbName := or.String(opts.DbName(), "gettr")
	uri := fmt.Sprintf("mongodb://localhost:%d", port)
	log.Printf("trying to connect to %s to create %s", uri, dbName)
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Printf("connected to %s", dbName)

	db := client.Database(dbName)
	db.Collection("userInfo")
	db.Collection("following")
	db.Collection("followers")

	res := &DB{
		dbName: dbName,
		client: client,
	}
	return res, nil
}

func (d *DB) database() *mongo.Database {
	return d.client.Database(d.dbName)
}

func (d *DB) collection(name string) *mongo.Collection {
	return d.database().Collection(name)
}

func (d *DB) Disconnect(ctx context.Context) error {
	return d.client.Disconnect(ctx)
}

var (
	dbVerboseUserInfo = flags.Bool("db_verbose_user_info", "verbose logging for getting and setting user info")
)

type UserOptions struct {
	Skip          bool
	FollowersDone bool
	FollowingDone bool
}

type storedUserInfo struct {
	UserInfo api.UserInfo
	Options  UserOptions
}

func (d *DB) SetUserInfo(ctx context.Context, username string, userInfo api.UserInfo) error {
	filter := bson.D{{"userinfo.username", username}}
	if res, err := d.collection("userInfo").DeleteMany(ctx, filter); err != nil {
		if *dbVerboseUserInfo {
			log.Printf("SetUserInfo: DeleteMany error: %v", err)
		}
		return err
	} else {
		if *dbVerboseUserInfo {
			log.Printf("SetUserInfo: DeleteMany result: %+v", res)
		}
	}

	stored := storedUserInfo{
		UserInfo: userInfo,
	}
	res, err := d.collection("userInfo").InsertOne(ctx, stored)
	if err != nil {
		return err
	}
	if *dbVerboseUserInfo {
		log.Printf("SetUserInfo(%q) -> %+v", username, res)
	}
	return nil
}

func (d *DB) SetUserOptions(ctx context.Context, username string, userOptions UserOptions) error {
	filter := bson.D{{"userinfo.username", username}}
	update := bson.D{
		{"$set", bson.D{
			{"options", userOptions},
		}},
	}
	if res, err := d.collection("userInfo").UpdateOne(ctx, filter, update); err == nil {
		if *dbVerboseFollowers {
			log.Printf("SetUserOptions for %s: UpdateOne result: %+v", username, res)
		}
		if res.MatchedCount != 0 || res.ModifiedCount != 0 || res.UpsertedCount != 0 {
			return nil
		}
	} else {
		if *dbVerboseFollowers {
			log.Printf("SetUserOptions for %s: UpdateOne error: %v", username, err)
		}
	}

	userInfo := api.UserInfo{
		Username: username,
	}
	stored := storedUserInfo{
		UserInfo: userInfo,
		Options:  userOptions,
	}
	res, err := d.collection("userInfo").InsertOne(ctx, stored)
	if err != nil {
		return err
	}
	if *dbVerboseUserInfo {
		log.Printf("SetUserInfo(%q) -> %+v", username, res)
	}
	return nil
}

func (d *DB) SetUserSkip(ctx context.Context, username string, skip bool) error {
	return d.setUserOptionsPart(ctx, username, skip, "skip", func() UserOptions {
		return UserOptions{
			Skip: skip,
		}
	})
}

func (d *DB) SetUserFollowersDone(ctx context.Context, username string, done bool) error {
	return d.setUserOptionsPart(ctx, username, done, "followersdone", func() UserOptions {
		return UserOptions{
			FollowersDone: done,
		}
	})
}

func (d *DB) SetUserFollowingDone(ctx context.Context, username string, done bool) error {
	return d.setUserOptionsPart(ctx, username, done, "followingdone", func() UserOptions {
		return UserOptions{
			FollowingDone: done,
		}
	})
}

func (d *DB) setUserOptionsPart(ctx context.Context, username string, val bool, part string, userOptionsCtor func() UserOptions) error {
	filter := bson.D{{"userinfo.username", username}}
	update := bson.D{
		{"$set", bson.D{
			{"options." + part, val},
		}},
	}
	if res, err := d.collection("userInfo").UpdateOne(ctx, filter, update); err == nil {
		if *dbVerboseFollowers {
			log.Printf("SetUserOptions for %s and part %s: UpdateOne result: %+v", username, part, res)
		}
		if res.MatchedCount != 0 || res.ModifiedCount != 0 || res.UpsertedCount != 0 {
			return nil
		}
	} else {
		if *dbVerboseFollowers {
			log.Printf("SetUserOptions for %s and part: UpdateOne error: %v", username, part, err)
		}
	}
	userInfo := api.UserInfo{
		Username: username,
	}
	stored := storedUserInfo{
		UserInfo: userInfo,
		Options:  userOptionsCtor(),
	}
	res, err := d.collection("userInfo").InsertOne(ctx, stored)
	if err != nil {
		return err
	}
	if *dbVerboseUserInfo {
		log.Printf("SetUserInfo(%q) -> %+v", username, res)
	}
	return nil
}

func (d *DB) getStoredUserInfo(ctx context.Context, username string) (*storedUserInfo, error) {
	filter := bson.D{{"userinfo.username", username}}
	res := &storedUserInfo{}
	if err := d.collection("userInfo").FindOne(ctx, filter).Decode(res); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DB) GetUserInfo(ctx context.Context, username string) (*api.UserInfo, error) {
	stored, err := d.getStoredUserInfo(ctx, username)
	if err != nil {
		return nil, err
	}
	return &stored.UserInfo, nil
}

func (d *DB) GetUserOptions(ctx context.Context, username string) (*UserOptions, error) {
	stored, err := d.getStoredUserInfo(ctx, username)
	if err != nil {
		return nil, err
	}
	return &stored.Options, nil
}

func noUsers(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no documents in result")
}

func (d *DB) GetUserSkip(ctx context.Context, username string) (bool, error) {
	userOptions, err := d.GetUserOptions(ctx, username)
	if noUsers(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return userOptions.Skip, nil
}

func (d *DB) GetUserFollowersDone(ctx context.Context, username string) (bool, error) {
	userOptions, err := d.GetUserOptions(ctx, username)
	if noUsers(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return userOptions.FollowersDone, nil
}

func (d *DB) GetUserFollowingDone(ctx context.Context, username string) (bool, error) {
	userOptions, err := d.GetUserOptions(ctx, username)
	if noUsers(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return userOptions.FollowingDone, nil
}

func (d *DB) deleteAllUserInfo(ctx context.Context) error {
	filter := bson.D{}
	if res, err := d.collection("userInfo").DeleteMany(ctx, filter); err != nil {
		if *dbVerboseUserInfo {
			log.Printf("deleteAllUserInfo: DeleteMany error: %v", err)
		}
		return err
	} else {
		if *dbVerboseUserInfo {
			log.Printf("deleteAllUserInfo: DeleteMany result: %+v", res)
		}
	}
	return nil
}

var (
	dbVerboseFollowers = flags.Bool("db_verbose_followers", "verbose logging for getting and setting followers")
	dbVerboseFollowing = flags.Bool("db_verbose_following", "verbose logging for getting and setting following")
)

type storedFollowish struct {
	Username  string
	Offset    int
	Usernames []string
}

func (d *DB) SetFollowers(ctx context.Context, username string, offset int, followers []string) error {
	return d.setFollowish(ctx, "followers", username, offset, followers)
}

func (d *DB) SetFollowing(ctx context.Context, username string, offset int, followers []string) error {
	return d.setFollowish(ctx, "following", username, offset, followers)
}

func (d *DB) setFollowish(ctx context.Context, collection, username string, offset int, usernames []string) error {
	filter := bson.D{
		{"username", username},
		{"offset", offset},
	}
	update := bson.D{
		{"$set", bson.D{
			{"usernames", usernames},
		}},
	}
	if res, err := d.collection(collection).UpdateOne(ctx, filter, update); err == nil {
		if *dbVerboseFollowers {
			log.Printf("setFollowish[%q]: UpdateOne result: %+v", collection, res)
		}
		if res.MatchedCount != 0 || res.ModifiedCount != 0 || res.UpsertedCount != 0 {
			return nil
		}
	} else {
		if *dbVerboseFollowers {
			log.Printf("setFollowish[%q]: UpdateOne error: %v", collection, err)
		}
	}

	insert := storedFollowish{
		Username:  username,
		Offset:    offset,
		Usernames: usernames,
	}
	res, err := d.collection(collection).InsertOne(ctx, insert)
	if err != nil {
		return err
	}
	if *dbVerboseFollowers {
		log.Printf("setFollowish[%q](%q) -> %+v", collection, username, res)
	}
	return nil
}

func (d *DB) deleteFollowers(ctx context.Context, username string) error {
	filter := bson.D{{"username", username}}
	if res, err := d.collection("followers").DeleteMany(ctx, filter); err != nil {
		if *dbVerboseFollowers {
			log.Printf("deleteFollowers: DeleteMany error: %v", err)
		}
		return err
	} else {
		if *dbVerboseFollowers {
			log.Printf("deleteFollowers: DeleteMany result: %+v", res)
		}
	}
	return nil
}

func (d *DB) deleteAllFollowers(ctx context.Context) error {
	filter := bson.D{}
	if res, err := d.collection("followers").DeleteMany(ctx, filter); err != nil {
		if *dbVerboseFollowers {
			log.Printf("deleteAllFollowers: DeleteMany error: %v", err)
		}
		return err
	} else {
		if *dbVerboseFollowers {
			log.Printf("deleteAllFollowers: DeleteMany result: %+v", res)
		}
	}
	return nil
}

func (d *DB) deleteAllFollowing(ctx context.Context) error {
	filter := bson.D{}
	if res, err := d.collection("following").DeleteMany(ctx, filter); err != nil {
		if *dbVerboseFollowing {
			log.Printf("deleteAllFollowing: DeleteMany error: %v", err)
		}
		return err
	} else {
		if *dbVerboseFollowing {
			log.Printf("deleteAllFollowing: DeleteMany result: %+v", res)
		}
	}
	return nil
}

func (d *DB) getFollowish(ctx context.Context, collection, username string) (chan string, chan error, error) {
	filter := bson.D{{"username", username}}
	findOpts := options.Find()
	findOpts.SetLimit(math.MaxInt)
	cur, err := d.collection(collection).Find(ctx, filter, findOpts)
	if err != nil {
		return nil, nil, errors.Errorf("%s Find: %v", collection, err)
	}

	followers := make(chan string)
	errors := make(chan error)
	go func() {
		for cur.Next(ctx) {
			var el storedFollowish
			if err := cur.Decode(&el); err != nil {
				errors <- err
				continue
			}
			for _, f := range el.Usernames {
				followers <- f
			}
		}
		close(followers)
		close(errors)
	}()
	return followers, errors, nil
}

func (d *DB) GetFollowers(ctx context.Context, username string) (chan string, chan error, error) {
	return d.getFollowish(ctx, "followers", username)
}

func (d *DB) GetFollowing(ctx context.Context, username string) (chan string, chan error, error) {
	return d.getFollowish(ctx, "following", username)
}

func (d *DB) GetFollowersSync(ctx context.Context, username string) ([]string, error) {
	filter := bson.D{{"username", username}}
	findOpts := options.Find()
	findOpts.SetLimit(math.MaxInt)
	cur, err := d.collection("followers").Find(ctx, filter, findOpts)
	if err != nil {
		return nil, errors.Errorf("Find: %v", err)
	}
	var res []string
	for cur.Next(ctx) {
		var el storedFollowish
		if err := cur.Decode(&el); err != nil {
			return nil, errors.Errorf("Decode: %v", err)
		}
		res = append(res, el.Usernames...)
	}
	return res, nil
}
