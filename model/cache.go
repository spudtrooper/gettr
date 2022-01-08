// A simple hash map implemented on disk.
package model

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	cacheVerbose = flag.Bool("cache_verbose", false, "verbose cache messages")
)

type Cache interface {
	Has(parts ...string) (bool, error)
	Set(parts ...string) error
	SetWithValue(val string, parts ...string) error
	SetBytes(val []byte, parts ...string) error
	SetInt(v int, parts ...string) error
	GetInt(parts ...string) (int, error)
	Get(parts ...string) ([]byte, error)
}

func MakeCache(dir string) Cache {
	if dir == "" {
		return &emptyCache{}
	}
	return &cacheImpl{
		dir: dir,
	}
}

func NonEmptyCache(c Cache) Cache {
	if c != nil {
		return c
	}
	return &emptyCache{}
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
	return c.SetWithValue("", parts...)
}

func (c *cacheImpl) SetWithValue(val string, parts ...string) error {
	return c.SetBytes([]byte(val), parts...)
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

func (c *cacheImpl) GetInt(parts ...string) (int, error) {
	b, err := c.get(parts...)
	if err != nil {
		return 0, nil
	}
	s := string(b)
	if s == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (c *cacheImpl) SetInt(v int, parts ...string) error {
	f := c.file(parts...)
	if err := c.writeFile(f, []byte(fmt.Sprintf("%d", v))); err != nil {
		return err
	}
	return nil
}

type emptyCache struct{}

func (c *emptyCache) Has(_ ...string) (bool, error) {
	return false, nil
}

func (c *emptyCache) Set(_ ...string) error {
	return nil
}

func (c *emptyCache) Get(_ ...string) ([]byte, error) {
	return nil, nil
}

func (c *emptyCache) SetWithValue(_ string, _ ...string) error {
	return nil
}

func (c *emptyCache) SetBytes(val []byte, parts ...string) error {
	return nil
}

func (c *emptyCache) GetInt(_ ...string) (int, error) {
	return 0, nil
}

func (c *emptyCache) SetInt(_ int, _ ...string) error {
	return nil
}
