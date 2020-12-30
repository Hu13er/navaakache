package cacheproxy

import (
	"io"
	"net/http"
)

type CacheHandler struct {
	Handler  http.Handler
	Cache    Cache
	TryCache func(req *http.Request) bool
}

var _ http.Handler = (*CacheHandler)(nil)

func (ch *CacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO
	if ch.tryCache(r) {
		if resp, err := ch.fetch(r); err == nil {
			w.WriteHeader(resp.StatusCode)
			for k, vs := range resp.Header {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			io.Copy(w, resp.Body)
			return
		}
		// Set cache here
		// ...
	}
	ch.Handler.ServeHTTP(w, r)
}

func (ch *CacheHandler) tryCache(req *http.Request) bool {
	return ch.TryCache != nil && ch.TryCache(req)
}

func (ch *CacheHandler) fetch(req *http.Request) (*http.Response, error) {
	if ch.Cache == nil {
		return nil, ErrorCacheNotFound
	}
	return ch.Cache.Fetch(req)
}

func (ch *CacheHandler) set(req *http.Request, resp http.Response) error {
	if ch.Cache == nil {
		return ErrorCacheCouldNotSave
	}
	return ch.Cache.Set(req, resp)
}
