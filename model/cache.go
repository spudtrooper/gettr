// A simple hash map implemented on disk.
package model

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

var (
	cacheVerbose = flag.Bool("cache_verbose", false, "verbose cache messages")
	cacheDir     = flag.String("cache_dir", "../gettrdata/data", "cache directory")
)

type Cache interface {
	Has(parts ...string) (bool, error)
	Set(parts ...string) error
	SetBytes(val []byte, parts ...string) error
	Get(parts ...string) ([]byte, error)
}

func MakeCacheFromFlags() (Cache, error) {
	if *cacheDir == "" {
		return nil, errors.Errorf("must set --cache_dir")
	}
	cache := makeCache(*cacheDir)
	return cache, nil

}

func makeCache(dir string) Cache {
	if dir == "" {
		return &emptyCache{}
	}
	return &cacheImpl{
		dir: dir,
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

func (c *cacheImpl) Get(parts ...string) ([]byte, error) {
	return c.get(parts...)
}

type emptyCache struct{}

func (c *emptyCache) Has(_ ...string) (bool, error)              { return false, nil }
func (c *emptyCache) Set(_ ...string) error                      { return nil }
func (c *emptyCache) Get(_ ...string) ([]byte, error)            { return nil, nil }
func (c *emptyCache) SetBytes(val []byte, parts ...string) error { return nil }
