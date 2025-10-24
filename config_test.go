package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Test non-embedded, disk based configuration.
func TestConfigNotEmbedded(t *testing.T) {

	var embeddedMode bool = false

	tests := []struct {
		name   string
		config string
		err    error
	}{
		{
			name: "ok",
			err:  nil,
			config: `
---
imageDir    : "images"
templateDir : "templates"
staticDir   : "static"
pageTemplate: "page.html"
indexTemplate: "index.html"
pages:
  -
    URL: "/home"
    Title: "Home"
    ImagePath: "images/home.jpg"
    Zones:
      -
        Left:   367
        Top:    44
        Right:  539
        Bottom: 263
        Target: "/detail"
  -
    URL: "/detail"
    Title: "Detail"
    ImagePath: "images/detail.jpg"
    Zones:
      -
        Left: 436
        Top:  31
        Right: 538
        Bottom: 73
        Target: "/home"
`},
		{
			name: "yaml parsing error",
			err:  errors.New("yaml parsing error"),
			config: `
---
pageTemplate: "page.html"
indexTemplate: "index.html"
pages:
    URL: "/home"
    Title: "Home"
    ImagePath: "images/home.jpg"
    Zones:
`},
		{
			name: "index template not set",
			err:  ErrInvalidConfig{"index template not set"},
			config: `
---
pageTemplate: "page.html"
pages:
  -
    URL: "/home"
    Title: "Home"
    ImagePath: "images/home.jpg"
    Zones:
      -
        Left:   367
        Top:    44
        Right:  539
        Bottom: 263
        Target: "/detail"
  -
    URL: "/detail"
    Title: "Detail"
    ImagePath: "images/detail.jpg"
    Zones:
      -
        Left: 436
        Top:  31
        Right: 538
        Bottom: 73
        Target: "/home"
`},

		{
			name: "too few pages",
			err:  ErrInvalidConfig{"too few pages"},
			config: `
---
pageTemplate: "page.html"
indexTemplate: "index.html"
pages:
  -
    URL: "/home"
    Title: "Home"
    ImagePath: "images/home.jpg"
    Zones:
      -
        Left:   367
        Top:    44
        Right:  539
        Bottom: 263
        Target: "/detail"
`},

		{
			name: "too few zones",
			err:  ErrInvalidConfig{"too few zones"},
			config: `
---
pageTemplate: "page.html"
indexTemplate: "index.html"
pages:
  -
    URL: "/home"
    Title: "Home"
    ImagePath: "images/home.jpg"
    Zones:
  -
    URL: "/detail"
    Title: "Detail"
    ImagePath: "images/detail.jpg"
    Zones:
      -
        Left: 436
        Top:  31
        Right: 538
        Bottom: 73
        Target: "/home"
`},
		{
			name: "invalid zone url",
			err:  ErrInvalidConfig{"invalid zone url"},
			config: `
---
pageTemplate: "page.html"
indexTemplate: "index.html"
pages:
  -
    URL: "/home"
    Title: "Home"
    ImagePath: "images/home.jpg"
    Zones:
      -
        Left:   367
        Top:    44
        Right:  539
        Bottom: 263
        Target: "/detail"
  -
    URL: "/detail"
    Title: "Detail"
    ImagePath: "images/detail.jpg"
    Zones:
      -
        Left: 436
        Top:  31
        Right: 538
        Bottom: 73
        Target: "/homes"
`},
		{
			name: "duplicate url",
			err:  ErrInvalidConfig{"duplicate url"},
			config: `
---
pageTemplate: "page.html"
indexTemplate: "index.html"
pages:
  -
    URL: "/home"
    Title: "Home"
    ImagePath: "images/home.jpg"
    Zones:
      -
        Left:   367
        Top:    44
        Right:  539
        Bottom: 263
        Target: "/detail"
  -
    URL: "/home"
    Title: "Detail"
    ImagePath: "images/detail.jpg"
    Zones:
      -
        Left: 436
        Top:  31
        Right: 538
        Bottom: 73
        Target: "/home"
`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newConfig([]byte(tt.config), embeddedMode)
			if tt.err == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error")
			}
			var expectedErr ErrInvalidConfig
			if errors.As(tt.err, &expectedErr) {
				var actualErr ErrInvalidConfig
				if !errors.As(err, &actualErr) {
					t.Fatalf("expected ErrInvalidConfig, got %T", err)
				}
			}
		})
	}
}

// Test embedded configuration.
func TestConfigEmbedded(t *testing.T) {

	var embeddedMode bool = true

	tests := []struct {
		name string
		err  error
	}{
		{name: "ok", err: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newConfig(configYaml, embeddedMode)
			if tt.err == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error")
			}
			var expectedErr ErrInvalidConfig
			if errors.As(tt.err, &expectedErr) {
				var actualErr ErrInvalidConfig
				if !errors.As(err, &actualErr) {
					t.Fatalf("expected ErrInvalidConfig, got %T", err)
				}
			}
		})
	}
}

func recursiveFSPrinter(t *testing.T, fi fs.FS) string {
	t.Helper()
	var s strings.Builder
	pathStrings := []string{"config.yaml", "images", "static", "templates"}
	err := fs.WalkDir(fi, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		pathParts := strings.Split(path, "/")
		for _, pp := range pathParts {
			for _, p := range pathStrings {
				if p == pp {
					if _, err := s.WriteString(fmt.Sprintf("%s\n", path)); err != nil {
						return fmt.Errorf("could not write %s", path)
					}
					return nil
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return s.String()
}

// Test writing embedded files to disk.
func TestConfigWriteEmbedded(t *testing.T) {

	want := recursiveFSPrinter(t, os.DirFS("."))

	dir, err := os.MkdirTemp("", "firstgo_embed_*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	c, err := newConfig(configYaml, true)
	if err != nil {
		t.Fatal(err)
	}
	err = c.validateConfig()
	if err != nil {
		t.Fatal(err)
	}

	err = c.WriteAssets(dir)
	if err != nil {
		t.Fatal(err)
	}

	got := recursiveFSPrinter(t, os.DirFS(dir))

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("got - want +: %v\n", diff)
	}

}
