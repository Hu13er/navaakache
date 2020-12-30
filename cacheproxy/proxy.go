package cacheproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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

	reverseProxy *httputil.ReverseProxy
	jsInjector   *JavascriptInjector
	cacher       *CacheHandler
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

	nc.jsInjector = (&JavascriptInjector{
		Handler: nc.reverseProxy,
	}).LoadFile("./cacheproxy/xhr_redefine.js")

	nc.cacher = &CacheHandler{
		Cache:    nopCacher{},
		Handler:  nc.jsInjector,
		TryCache: nc.tryCache,
	}

	return nc, nil
}

func (nc *navaakCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nc.cacher.ServeHTTP(w, r)
}

func (nc *navaakCache) reverseProxyDirector(req *http.Request) {
	host := req.Host
	if len(host) <= 0 {
		host = req.URL.Host
	}
	req.Host = ""
	req.URL.Scheme = nc.NavaakURL.Scheme
	parts := strings.Split(host, ".")
	parts[len(parts)-1] = nc.NavaakURL.Host
	req.URL.Host = strings.Join(parts, ".")
	// TODO: already we dont accept encoding (e.g. gzip)
	// and our js injector expect plain stream.
	req.Header.Del("Accept-Encoding")
}

func (nc *navaakCache) tryCache(req *http.Request) bool {
	// Try to cache stream.navaak.com
	return req.URL.Host == nc.NavaakStreamURL.Host
}
