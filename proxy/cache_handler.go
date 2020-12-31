package proxy

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

type cacheHandler struct {
	handler       http.Handler
	requestCacher requestCacher
	tryForCache   func(req *http.Request) bool
}

var _ http.Handler = (*cacheHandler)(nil)

func (ch *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: handle errors.
	if ch.tryCache(r) {
		if resp, err := ch.fetch(r); err == nil {
			for k, vs := range resp.Header {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
			return
		}

		resp := newResponseRecorder()
		ch.handler.ServeHTTP(resp, r)
		body := resp.Buff.Bytes()
		for k, vs := range resp.HeaderMap {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.Code)
		w.Write(body)

		ch.set(r, http.Response{
			StatusCode: resp.Code,
			Header:     resp.HeaderMap,
			Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
		})

		return
	}
	ch.handler.ServeHTTP(w, r)
}

func (ch *cacheHandler) tryCache(req *http.Request) bool {
	return ch.tryForCache != nil && ch.tryForCache(req)
}

func (ch *cacheHandler) fetch(req *http.Request) (*http.Response, error) {
	if ch.requestCacher == nil {
		return nil, errorCacheNotFound
	}
	return ch.requestCacher.fetch(req)
}

func (ch *cacheHandler) set(req *http.Request, resp http.Response) error {
	if ch.requestCacher == nil {
		return errorCacheCouldNotSave
	}
	return ch.requestCacher.set(req, resp)
}
