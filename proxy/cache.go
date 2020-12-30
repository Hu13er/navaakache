package proxy

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	errorCacheNotFound     = errors.New("error cache not found")
	errorCacheCouldNotSave = errors.New("error cache couldnt save")
)

type requestCacher interface {
	fetch(*http.Request) (*http.Response, error)
	set(*http.Request, http.Response) error
}

type nopCacher struct{}

var _ requestCacher = nopCacher{}

func (nop nopCacher) fetch(req *http.Request) (*http.Response, error) {
	fmt.Println("Fetch", req)
	return nil, errorCacheNotFound
}

func (nop nopCacher) set(req *http.Request, resp http.Response) error {
	fmt.Println("Set", req, "to", resp)
	return errorCacheCouldNotSave
}
