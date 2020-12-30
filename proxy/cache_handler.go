package proxy

import (
	"io"
	"net/http"
)

type cacheHandler struct {
	handler     http.Handler
	cache       requestCacher
	tryForCache func(req *http.Request) bool
}

var _ http.Handler = (*cacheHandler)(nil)

func (ch *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	ch.handler.ServeHTTP(w, r)
}

func (ch *cacheHandler) tryCache(req *http.Request) bool {
	return ch.tryForCache != nil && ch.tryForCache(req)
}

func (ch *cacheHandler) fetch(req *http.Request) (*http.Response, error) {
	if ch.cache == nil {
		return nil, errorCacheNotFound
	}
	return ch.cache.fetch(req)
}

func (ch *cacheHandler) set(req *http.Request, resp http.Response) error {
	if ch.cache == nil {
		return errorCacheCouldNotSave
	}
	return ch.cache.set(req, resp)
}
