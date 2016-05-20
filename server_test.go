package main

import (
	"bytes"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var statusesTest = []struct {
	method string
	path   string
	body   []byte
	code   int
}{
	{"HEAD", "/", nil, 405},
	{"PUT", "/", nil, 405},
	{"GET", "/foo", nil, 404},
	{"POST", "/", nil, 400},          // No body
	{"POST", "/", []byte("{}"), 400}, // Wrong body
	{"POST", "/", []byte(
		`{"coinbase": "0x1111111111111111111111111111111111111111"}`), 201},
}

func TestStatuses(t *testing.T) {
	for _, tt := range statusesTest {
		r, _ := http.NewRequest(
			tt.method,
			"https://ethercubi.lol/"+tt.path,
			bytes.NewBuffer(tt.body),
		)

		h := cubiHandler()
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		if rr.Code != tt.code {
			t.Errorf("%s %s = %d want %d", tt.method, tt.path, rr.Code, tt.code)
		}
	}
}

func TestSampleSize(t *testing.T) {
	list := []string{"a", "b", "c", "d"}
	s, _ := sample(list, 2)
	if len(s) != 2 {
		t.Errorf("sample length : %s want %d", len(s), 2)
	}
}

func TestSampleInvalidSize(t *testing.T) {
	list := []string{"a", "b", "c", "d"}
	_, err := sample(list, 5)
	if err == nil {
		t.Error("expected an invalid sample size error")
	}
}

var addressTest = []struct {
	addr  string
	valid bool
}{
	{"", false},
	{"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", false},
	{"0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", true},
}

func TestValidAddress(t *testing.T) {
	for _, tt := range addressTest {
		r := validAddress(tt.addr)
		if r != tt.valid {
			t.Errorf("%s valid: %t wants %t", tt.addr, r, tt.valid)
		}
	}
}

func TestPostGet(t *testing.T) {
	post, _ := http.NewRequest(
		"POST",
		"https://ethercubi.lol/",
		bytes.NewBuffer([]byte(
			`{"coinbase": "0x1111111111111111111111111111111111111111"}`)),
	)

	h := cubiHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, post)
	if rr.Code != 201 {
		t.Errorf("POST returned %d wants 201", rr.Code)
	}

	t.Log(rr.HeaderMap)
	t.Log(rr.Body)

}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	retCode := m.Run()
	os.Exit(retCode)
}
