package main

import (
	"errors"
	"testing"
)

func TestConfig(t *testing.T) {

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
pageTemplate: "templates/page.html"
indexTemplate: "templates/index.html"
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
pageTemplate: "templates/page.html"
indexTemplate: "templates/index.html"
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
pageTemplate: "templates/page.html"
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
pageTemplate: "templates/page.html"
indexTemplate: "templates/index.html"
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
pageTemplate: "templates/page.html"
indexTemplate: "templates/index.html"
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
pageTemplate: "templates/page.html"
indexTemplate: "templates/index.html"
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
pageTemplate: "templates/page.html"
indexTemplate: "templates/index.html"
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
			_, err := newConfig([]byte(tt.config))
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
