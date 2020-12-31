package proxy

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Hu13er/navaakache/cache"
)

var (
	errorCacheNotFound     = errors.New("error cache not found")
	errorCacheCouldNotSave = errors.New("error cache couldnt save")
)

type requestCacher interface {
	fetch(*http.Request) (*http.Response, error)
	set(*http.Request, http.Response) error
}

type streamCacher struct {
	backend cache.Cacher
}

var _ requestCacher = (*streamCacher)(nil)

func (sc *streamCacher) set(req *http.Request, resp http.Response) error {
	key, err := sc.keyof(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	val, err := responseEncodeDecoder{
		Code:   resp.StatusCode,
		Header: resp.Header,
		Body:   body,
	}.encode()
	if err != nil {
		return err
	}
	return sc.backend.Set(key, val)
}

func (sc *streamCacher) fetch(req *http.Request) (*http.Response, error) {
	key, err := sc.keyof(req)
	if err != nil {
		return nil, err
	}
	encoded, err := sc.backend.Fetch(key)
	if err != nil {
		return nil, err
	}
	var val responseEncodeDecoder
	if err := (&val).decode(encoded); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: val.Code,
		Header:     val.Header,
		Body:       ioutil.NopCloser(bytes.NewBuffer(val.Body)),
	}, nil
}

func (sc *streamCacher) keyof(req *http.Request) (string, error) {
	r := regexp.MustCompile(`\/aes\/_definst_\/smil:(.*?)\/(.*?)\/media_.*?_(.*?)\.aac`)
	matches := r.FindStringSubmatch(req.URL.String())
	if len(matches) < 4 {
		return "", errors.New("not a streaming url")
	}
	return strings.Join(matches[1:], "@"), nil
}

type responseEncodeDecoder struct {
	Code   int
	Header http.Header
	Body   []byte
}

func (r responseEncodeDecoder) encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *responseEncodeDecoder) decode(encoded []byte) error {
	buf := bytes.NewBuffer(encoded)
	return gob.NewDecoder(buf).Decode(r)
}

type nopCacher struct{}

var _ requestCacher = nopCacher{}

func (nop nopCacher) fetch(req *http.Request) (*http.Response, error) {
	fmt.Println("Fetch", req)
	return nil, errorCacheNotFound
}

func (nop nopCacher) set(req *http.Request, resp http.Response) error {
	fmt.Println("Set", req, "to", resp)
	return errorCacheCouldNotSave
}
