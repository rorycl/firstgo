package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func initServer() (*server, error) {
	s, err := newServer(
		"127.0.0.1",
		"8001",
		[]page{
			page{"/home", "Home", "images/home.jpg", []pageZone{}},
			page{"/detail", "Detail", "images/detail.jpg", []pageZone{}},
		},
		"templates/page.html",
	)
	return s, err
}

func TestNewServer(t *testing.T) {
	_, err := initServer()
	if err != nil {
		t.Fatal(err)
	}
}

func TestServerFavicon(t *testing.T) {
	s, err := initServer()
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodGet, "http://example.com/favicon.ico", nil)
	w := httptest.NewRecorder()

	s.Favicon(w, r)

	res := w.Result()
	defer res.Body.Close()
	_, err = io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if want, got := 200, res.StatusCode; want != got {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

func TestServerHealth(t *testing.T) {
	s, err := initServer()
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodGet, "http://example.com/health", nil)
	w := httptest.NewRecorder()

	s.Health(w, r)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if want, got := 200, res.StatusCode; want != got {
		t.Errorf("expected status %d, got %d", want, got)
	}
	responseBody := string(data)
	if want, got := strings.TrimSpace(`{"status":"up"}`), strings.TrimSpace(responseBody); want != got {
		t.Errorf("expected status %s, got %s", want, got)
	}
}

func TestServerServe(t *testing.T) {
	s, err := initServer()
	if err != nil {
		t.Fatal(err)
	}
	testServer := &httptest.Server{}
}
