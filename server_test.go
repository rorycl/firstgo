package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// initServer inits a server with default content in the repo, such as
// the material at assets/static (including styles) and pages (including
// images from assets/images and templates from assets/templates).
func initServer(t *testing.T) *server {
	t.Helper()
	cfg := &config{
		AssetsDir:     "assets",
		PageTemplate:  "templates/page.html",
		IndexTemplate: "templates/index.html",
		Pages: []page{
			page{
				URL:       "/home",
				Title:     "Home",
				ImagePath: "images/home.jpg",
				// Note:      "",
				Zones: []pageZone{pageZone{367, 44, 539, 263, "/detail", ""}},
			},
			page{
				URL:       "/detail",
				Title:     "Detail",
				ImagePath: "images/detail.jpg",
				Note:      "",
				Zones:     []pageZone{pageZone{436, 31, 538, 73, "/home", ""}},
			},
		},
	}
	if err := cfg.validateConfig(); err != nil {
		t.Fatal(err)
	}
	s, err := newServer(
		"127.0.0.1",
		"8001",
		cfg,
	)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestServer tests a running server instance of the site using the
// configuration loaded from initServer. The configuration depends on
// the on-disk default content in the repo.
func TestServer(t *testing.T) {

	s := initServer(t)

	handler, err := s.buildHandler()
	if err != nil {
		t.Fatal("buildHander error:", err)
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
		{"Favicon ico", "/favicon.ico", http.StatusOK, "<svg xmlns="},
		{"Image File", "/images/home.jpg", http.StatusOK, "Photoshop 3.0"},
		{"Index", "/index", http.StatusOK, "<h1>Index</h1>"},
		{"Root", "/", http.StatusOK, "<h1>Index</h1>"},
		{"Not Found", "/nonexistent", http.StatusNotFound, "404 page not found"},
		{"Templates file 404", "/templates", http.StatusNotFound, "404 page not found"},
		{"Templates dir 404", "/templates/", http.StatusNotFound, "The templates directory is purposely not mounted."},
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
