package proxy

import (
	"bytes"
	"io"
	"net/http"
)

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
