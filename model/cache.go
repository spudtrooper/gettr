// A simple hash map implemented on disk.
package model

import (
	"encoding/json"
	"flag"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/log"
)

var (
	cacheVerbose = flag.Bool("cache_verbose", false, "verbose cache messages")
	cacheDir     = flag.String("cache_dir", "../gettrdata/data", "cache directory")
)

type Cache interface {
	Has(parts ...string) (bool, error)
	Set(parts ...string) error
	SetBytes(val []byte, parts ...string) error
	GetBytes(parts ...string) ([]byte, error)
	SetGeneric(val interface{}, parts ...string) error
	GetStrings(parts ...string) ([]string, error)
	GetAllStrings(parts ...string) (SharedStrings, error)
	FindKeys(parts ...string) ([]string, error)
	FindKeysChannels(parts ...string) (chan string, chan error, error)
}

func MakeCacheFromFlags() (Cache, error) {
	if *cacheDir == "" {
		return nil, errors.Errorf("must set --cache_dir")
	}
	cache := makeCache(*cacheDir)
	return cache, nil
}

func makeCache(dir string) *cacheImpl {
	return &cacheImpl{
		dir: *cacheDir,
	}
}

type cacheImpl struct {
	dir string
}

func (c *cacheImpl) file(parts ...string) string {
	key := strings.Join(parts, "/")
	file := path.Join(c.dir, key)
	return file
}

func (c *cacheImpl) fileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *cacheImpl) isDir(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if os.IsNotExist(err) {
		log.Printf("%s doesn't exit", filename)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func (c *cacheImpl) writeFile(f string, b []byte) error {
	if err := os.MkdirAll(path.Dir(f), 0755); err != nil {
		return err
	}
	if *cacheVerbose {
		log.Printf("writing %d bytes to %s", len(b), f)
	}
	if err := ioutil.WriteFile(f, b, 0755); err != nil {
		return err
	}
	return nil
}

func (c *cacheImpl) Has(parts ...string) (bool, error) {
	f := c.file(parts...)
	exists, err := c.fileExists(f)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *cacheImpl) Set(parts ...string) error {
	return c.SetBytes(nil, parts...)
}

func (c *cacheImpl) SetBytes(val []byte, parts ...string) error {
	f := c.file(parts...)
	if err := c.writeFile(f, val); err != nil {
		return err
	}
	return nil
}

func (c *cacheImpl) get(parts ...string) ([]byte, error) {
	f := c.file(parts...)
	exists, err := c.fileExists(f)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	if *cacheVerbose {
		log.Printf("read %d bytes from %s", len(b), f)
	}
	return b, nil
}

func (c *cacheImpl) GetBytes(parts ...string) ([]byte, error) {
	return c.get(parts...)
}

func (c *cacheImpl) SetGeneric(val interface{}, parts ...string) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.SetBytes(bytes, parts...)
}

func (c *cacheImpl) GetStrings(parts ...string) ([]string, error) {
	bytes, err := c.get(parts...)
	if err != nil {
		return nil, err
	}
	var res []string
	if err := json.Unmarshal(bytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// GetAllStrings returns all []string in a directory
// Example:
//   we have
//     users
//       foo
//         followersOffsets
//              1 = [1,2,3]
//              2 = [4,5,6]
//              3 = [7,8,9]
//   GetAllStrings("user", "foo", "followersOffsets") == [1,2,3,4,5,6,7,8,9]
type SharedString struct{ Val, Dir string }

type SharedStrings []SharedString

func (s SharedStrings) Len() int           { return len(s) }
func (s SharedStrings) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SharedStrings) Less(i, j int) bool { return s[i].Val < s[j].Val }

func (s SharedStrings) Strings() []string {
	var res []string
	for _, x := range s {
		res = append(res, x.Val)
	}
	return res
}

func (c *cacheImpl) GetAllStrings(parts ...string) (SharedStrings, error) {
	dir := c.file(parts...)
	isDir, err := c.isDir(dir)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, nil
	}
	set := map[string]string{}
	if err := filepath.WalkDir(dir, func(path string, di fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if di.IsDir() {
			return nil
		}
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var arr []string
		if err := json.Unmarshal(bytes, &arr); err != nil {
			return err
		}
		for _, s := range arr {
			set[s] = filepath.Base(filepath.Dir(path))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	var res SharedStrings
	for s, o := range set {
		res = append(res, SharedString{
			Val: s,
			Dir: o,
		})
	}
	return res, nil
}

func (c *cacheImpl) FindKeys(parts ...string) ([]string, error) {
	dir := c.file(parts...)
	isDir, err := c.isDir(dir)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, errors.Errorf("%s is not a directory", dir)
	}
	var res []string
	if err := filepath.WalkDir(dir, func(path string, di fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if di.IsDir() {
			return nil
		}
		res = append(res, filepath.Base(path))
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: This just looks do directories
func (c *cacheImpl) FindKeysChannels(parts ...string) (chan string, chan error, error) {
	keys := make(chan string)
	errs := make(chan error)

	dir := c.file(parts...)
	isDir, err := c.isDir(dir)
	if err != nil {
		return nil, nil, err
	}
	if !isDir {
		return nil, nil, errors.Errorf("%s is not a directory", dir)
	}

	go func() {
		if err := filepath.WalkDir(dir, func(path string, di fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !di.IsDir() {
				return nil
			}
			if path == dir {
				return nil
			}
			key := strings.Replace(path, dir+"/", "", 1)
			if key == "" {
				return nil
			}
			keys <- key
			return nil
		}); err != nil {
			errs <- err
		}
		close(keys)
		close(errs)
	}()

	return keys, errs, nil
}
