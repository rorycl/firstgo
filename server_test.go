package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func initServer(t *testing.T) *server {
	s, err := newServer(
		"127.0.0.1",
		"8001",
		[]page{
			page{"/home", "Home", "images/home.jpg", []pageZone{
				pageZone{367, 44, 539, 263, "/detail"},
			}},
			page{"/detail", "Detail", "images/detail.jpg", []pageZone{
				pageZone{436, 31, 538, 73, "/home"},
			}},
		},
		"templates/page.html",
	)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestServer(t *testing.T) {

	s := initServer(t)

	handler, err := s.buildHandler()
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(handler)
	defer ts.Close()

	testCases := []struct {
		name         string
		path         string
		statusCode   int
		bodyContains string
	}{
		{"Health Check", "/health", http.StatusOK, `{"status":"up"}`},
		{"Home Page", "/home", http.StatusOK, "<title>Home"},
		{"Detail Page", "/detail", http.StatusOK, "<title>Detail"},
		{"Favicon", "/favicon", http.StatusOK, "<svg xmlns="},
		{"Image File", "/images/home.jpg", http.StatusOK, "Photoshop 3.0"},
		{"Not Found", "/nonexistent", http.StatusNotFound, "404 page not found"},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			client := ts.Client()
			url, err := url.JoinPath(ts.URL, tt.path)
			if err != nil {
				t.Fatalf("url joining error: %s, %s : %v", ts.URL, tt.path, err)
			}

			resp, err := client.Get(url)
			if err != nil {
				t.Fatalf("get error: %v", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if got, want := resp.StatusCode, tt.statusCode; got != want {
				t.Fatalf("got %d want %d", got, want)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("could not read body: %v", err)
			}
			if !bytes.Contains(body, []byte(tt.bodyContains)) {
				t.Errorf("body does not contain %q", tt.bodyContains)
			}
		})
	}
}
