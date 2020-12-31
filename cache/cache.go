package cache

import (
	"database/sql"
	"encoding/hex"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/faabiosr/cachego"
	"github.com/faabiosr/cachego/sqlite3"
)

const (
	CacheMaxDuration = 6 * 30 * 24 * time.Hour // 6months
)

type Cacher interface {
	Set(key string, value []byte) error
	Fetch(key string) ([]byte, error)
}

type cacheGo struct {
	backend cachego.Cache
}

var _ Cacher = (*cacheGo)(nil)

func NewCacheGo() (*cacheGo, error) {
	db, err := sql.Open("sqlite3", "./cache.db")
	if err != nil {
		return nil, err
	}
	cache, err := sqlite3.New(db, "cache")
	if err != nil {
		return nil, err
	}
	return &cacheGo{
		backend: cache,
	}, nil
}

func (c *cacheGo) Set(key string, value []byte) error {
	encoded := hex.EncodeToString(value)
	return c.backend.Save(key, encoded, CacheMaxDuration)
}

func (c *cacheGo) Fetch(key string) ([]byte, error) {
	encoded, err := c.backend.Fetch(key)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(encoded)
}
