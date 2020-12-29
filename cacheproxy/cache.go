package cacheproxy

import (
	"errors"
	"net/http"
)

var ErrorCacheNotFound = errors.New("error cache not found")
var ErrorCacheCouldNotSave = errors.New("error cache couldnt save")

type Cache interface {
	Fetch(*http.Request) (*http.Response, error)
	Set(*http.Request, http.Response) error
}

type nopCacher struct{}

var _ Cache = nopCacher{}

func (nop nopCacher) Fetch(*http.Request) (*http.Response, error) {
	return nil, ErrorCacheNotFound
}

func (nop nopCacher) Set(*http.Request, http.Response) error {
	return ErrorCacheCouldNotSave
}
