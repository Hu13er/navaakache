package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Hu13er/navaakache/cache"
)

var DefaultNavaakCache *navaakCache

func init() {
	// var err error
	c, err := cache.NewCacheGo()
	if err != nil {
		panic(err)
	}
	DefaultNavaakCache, err = NewNavaakCache(Configs{
		LocalAddr: "localhost:8000",
		URL:       "https://navaak.com",
		Stream:    "https://stream.navaak.com",
		Cacher:    c,
	})
	if err != nil {
		panic(err)
	}
}

type Configs struct {
	LocalAddr string
	URL       string
	Stream    string
	Cacher    cache.Cacher
}

type navaakCache struct {
	LocalAddr       string
	NavaakURL       url.URL
	NavaakStreamURL url.URL
	Cacher          cache.Cacher

	reverseProxy  *httputil.ReverseProxy
	jsInjector    *javascriptInjector
	cacherHandler *cacheHandler
}

func NewNavaakCache(confs Configs) (*navaakCache, error) {
	nc := &navaakCache{}

	nc.LocalAddr = confs.LocalAddr

	u, err := url.Parse(confs.URL)
	if err != nil {
		return nil, err
	}
	nc.NavaakURL = *u

	u, err = url.Parse(confs.URL)
	if err != nil {
		return nil, err
	}
	nc.NavaakStreamURL = *u

	nc.Cacher = confs.Cacher

	nc.reverseProxy = &httputil.ReverseProxy{
		Director: nc.reverseProxyDirector,
	}

	nc.jsInjector = (&javascriptInjector{
		Handler: nc.reverseProxy,
	}).loadFile("./proxy/xhr_redefine.js", map[string]string{
		"PROXY_ADDR": "'" + nc.LocalAddr + "'",
	})

	nc.cacherHandler = &cacheHandler{
		requestCacher: &streamCacher{
			backend: nc.Cacher,
		},
		handler:     nc.jsInjector,
		tryForCache: nc.tryCache,
	}

	return nc, nil
}

func (nc *navaakCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nc.cacherHandler.ServeHTTP(w, r)
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
	// Try to cache stream.localhost:8000
	host := req.Host
	if len(host) <= 0 {
		host = req.URL.Host
	}
	return strings.HasPrefix(host, "stream.")
}
