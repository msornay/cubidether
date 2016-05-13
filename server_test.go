package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var clientErrorTest = []struct {
	method string
	path   string
	code   int
}{
	{"HEAD", "/", 405},
	{"PUT", "/", 405},
	{"GET", "/XXX", 404},
}

func TestClientError(t *testing.T) {
	for _, tt := range clientErrorTest {
		r := &http.Request{
			Method: tt.method,
			URL: &url.URL{
				Path: tt.path,
			},
		}
		h := cubiHandler()
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		if rr.Code != tt.code {
			t.Errorf("%s %s = %d want %d", tt.method, tt.path, rr.Code, tt.code)
		}
	}
}
