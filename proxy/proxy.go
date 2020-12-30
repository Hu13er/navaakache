package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var DefaultNavaakCache *navaakCache

func init() {
	var err error
	DefaultNavaakCache, err = NewNavaakCache(Configs{
		LocalAddr: "localhost:8000",
		URL:       "https://navaak.com",
		Stream:    "https://stream.navaak.com",
	})
	if err != nil {
		panic(err)
	}
}

type Configs struct {
	LocalAddr string
	URL       string
	Stream    string
}

type navaakCache struct {
	LocalAddr       string
	NavaakURL       url.URL
	NavaakStreamURL url.URL

	reverseProxy *httputil.ReverseProxy
	jsInjector   *javascriptInjector
	cacher       *cacheHandler
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

	nc.reverseProxy = &httputil.ReverseProxy{
		Director: nc.reverseProxyDirector,
	}

	nc.jsInjector = (&javascriptInjector{
		Handler: nc.reverseProxy,
	}).loadFile("./proxy/xhr_redefine.js", map[string]string{
		"PROXY_ADDR": "'" + nc.LocalAddr + "'",
	})

	nc.cacher = &cacheHandler{
		cache:       nopCacher{},
		handler:     nc.jsInjector,
		tryForCache: nc.tryCache,
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
