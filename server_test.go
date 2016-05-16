package main

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

var statusesTest = []struct {
	method string
	path   string
	code   int
}{
	{"HEAD", "/", 405},
	{"PUT", "/", 405},
	{"GET", "/XXX", 404},
	{"POST", "/", 200},
}

func TestStatuses(t *testing.T) {
	for _, tt := range statusesTest {
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


func TestSample(t *testing.T) {
	list := []string{"a", "b", "c", "d"}
	s := sample(list, 2)
	if len(s) != 2 {
		t.Errorf("sample length : %s want %d", len(s), 2)
	}
}


func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	retCode := m.Run()
	os.Exit(retCode)
}
