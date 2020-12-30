package proxy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type javascriptInjector struct {
	Handler http.Handler
	JSCode  string
	Only    func(req *http.Request) bool
}

var _ http.Handler = (*javascriptInjector)(nil)

func (jsi *javascriptInjector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := newResponseRecorder()
	jsi.Handler.ServeHTTP(resp, r)

	if jsi.Only != nil && !jsi.Only(r) {
		resp.writeResponse(w)
		return
	}

	var isTextHTML bool
	contentType := resp.Header().Values("Content-Type")
	for _, v := range contentType {
		v = strings.ToLower(v)
		if strings.Contains(v, "text/html") {
			isTextHTML = true
		}
	}
	if !isTextHTML {
		resp.writeResponse(w)
		return
	}

	html := resp.Buff.String()
	parts := strings.SplitN(html, "<head>", 2)
	if len(parts) <= 1 {
		resp.writeResponse(w)
		return
	}
	newHTMLFmt := `%s
<head>
<!-- Injected script by Navaakache -->
<script>
%s
</script>
%s`
	newHTML := fmt.Sprintf(newHTMLFmt, parts[0], jsi.JSCode, parts[1])
	resp.Buff = bytes.NewBufferString(newHTML)
	resp.writeResponse(w)
}

func (jsi *javascriptInjector) loadFile(filename string) *javascriptInjector {
	js, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	jsi.JSCode = string(js)
	return jsi
}

type responseRecorder struct {
	Code      int
	HeaderMap http.Header
	Buff      *bytes.Buffer
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{
		Code:      200,
		HeaderMap: make(http.Header),
		Buff:      &bytes.Buffer{},
	}
}

var _ http.ResponseWriter = (*responseRecorder)(nil)

func (rr *responseRecorder) writeResponse(w http.ResponseWriter) {
	for k, vs := range rr.HeaderMap {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(rr.Code)
	io.Copy(w, rr.Buff)
}

func (rr *responseRecorder) Header() http.Header {
	m := rr.HeaderMap
	if m == nil {
		m = make(http.Header)
		rr.HeaderMap = m
	}
	return m
}

func (rr *responseRecorder) Write(buf []byte) (int, error) {
	return rr.Buff.Write(buf)
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.Code = statusCode
}
