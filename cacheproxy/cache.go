package cacheproxy

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrorCacheNotFound     = errors.New("error cache not found")
	ErrorCacheCouldNotSave = errors.New("error cache couldnt save")
)

type Cache interface {
	Fetch(*http.Request) (*http.Response, error)
	Set(*http.Request, http.Response) error
}

type nopCacher struct{}

var _ Cache = nopCacher{}

func (nop nopCacher) Fetch(req *http.Request) (*http.Response, error) {
	fmt.Println("Fetch", req)
	return nil, ErrorCacheNotFound
}

func (nop nopCacher) Set(req *http.Request, resp http.Response) error {
	fmt.Println("Set", req, "to", resp)
	return ErrorCacheCouldNotSave
}
