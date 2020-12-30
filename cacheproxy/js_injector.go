package cacheproxy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type JavascriptInjector struct {
	Handler http.Handler
	JSCode  string
	Only    func(req *http.Request) bool
}

var _ http.Handler = (*JavascriptInjector)(nil)

func (jsi *JavascriptInjector) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fmt.Println("REQ header:", r.Header)
	resp := NewResponseRecorder()
	jsi.Handler.ServeHTTP(resp, r)
	fmt.Println("Resp Header:", resp.Header())

	if jsi.Only != nil && !jsi.Only(r) {
		resp.WriteResponse(w)
		return
	}

	var isTextHTML bool
	contentType := resp.Header().Values("Content-Type")
	fmt.Println("CONTENT-TYPE", contentType)
	for _, v := range contentType {
		v = strings.ToLower(v)
		if strings.Contains(v, "text/html") {
			fmt.Println("HERE!")
			isTextHTML = true
		}
	}
	fmt.Println("ISHTML", isTextHTML)
	if !isTextHTML {
		resp.WriteResponse(w)
		return
	}

	html := resp.Buff.String()
	parts := strings.SplitN(html, "<head>", 2)
	if len(parts) <= 1 {
		resp.WriteResponse(w)
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
	resp.WriteResponse(w)
}

func (jsi *JavascriptInjector) LoadFile(filename string) *JavascriptInjector {
	js, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	jsi.JSCode = string(js)
	return jsi
}

type ResponseRecorder struct {
	Code      int
	HeaderMap http.Header
	Buff      *bytes.Buffer
}

func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		Code:      200,
		HeaderMap: make(http.Header),
		Buff:      &bytes.Buffer{},
	}
}

var _ http.ResponseWriter = (*ResponseRecorder)(nil)

func (rr *ResponseRecorder) WriteResponse(w http.ResponseWriter) {
	for k, vs := range rr.HeaderMap {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(rr.Code)
	io.Copy(w, rr.Buff)
}

func (rr *ResponseRecorder) Header() http.Header {
	m := rr.HeaderMap
	if m == nil {
		m = make(http.Header)
		rr.HeaderMap = m
	}
	return m
}

func (rr *ResponseRecorder) Write(buf []byte) (int, error) {
	return rr.Buff.Write(buf)
}

func (rr *ResponseRecorder) WriteHeader(statusCode int) {
	rr.Code = statusCode
}
