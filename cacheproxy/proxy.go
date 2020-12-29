package cacheproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

var DefaultNavaakCache *navaakCache

func init() {
	var err error
	DefaultNavaakCache, err = NewNavaakCache(NavaakURLs{
		Default: "https://navaak.com",
		Stream:  "https://stream.navaak.com",
	})
	if err != nil {
		panic(err)
	}
}

type NavaakURLs struct {
	Default string
	Stream  string
}

type navaakCache struct {
	NavaakURL       url.URL
	NavaakStreamURL url.URL
	reverseProxy    *httputil.ReverseProxy
	cacher          *CacheHandler
}

func NewNavaakCache(urls NavaakURLs) (*navaakCache, error) {
	nc := &navaakCache{}

	u, err := url.Parse(urls.Default)
	if err != nil {
		return nil, err
	}
	nc.NavaakURL = *u

	u, err = url.Parse(urls.Default)
	if err != nil {
		return nil, err
	}
	nc.NavaakStreamURL = *u

	nc.reverseProxy = &httputil.ReverseProxy{
		Director: nc.reverseProxyDirector,
	}
	nc.cacher = &CacheHandler{
		Handler:  nc.reverseProxy,
		TryCache: nc.tryCache,
	}
	return nc, nil
}

func (nc *navaakCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nc.cacher.ServeHTTP(w, r)
}

func (nc *navaakCache) reverseProxyDirector(req *http.Request) {
	req.Host = ""
	req.URL.Scheme = nc.NavaakURL.Scheme
	req.URL.Host = nc.NavaakURL.Host
}

func (nc *navaakCache) tryCache(*http.Request) bool {
	return false
}
