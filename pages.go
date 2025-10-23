package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/goccy/go-yaml"

	"text/template"
)

// config describes the config in a yaml configuration file
type config struct {
	PageTemplate string `yaml:"pageTemplate"`
	Pages        []page `yaml:"pages"`
}

// newConfig creates and validates a new config from reading a yaml
// file.
func newConfig(path string) (*config, error) {
	r, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config '%q' reading error: %v", err)
	}
	var c config
	if err = yaml.Unmarshal(r, &c); err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}
	return &c, nil
}

// pageZone sets up a rectangular page zone on a page that, when
// clicked, redirects to Target.
type pageZone struct {
	Left   int    `yaml:"Left"`
	Top    int    `yaml:"Top"`
	Right  int    `yaml:"Right"`
	Bottom int    `yaml:"Bottom"`
	Target string `yaml:"Target"`
}

// Width returns the width of the pageZone.
func (p *pageZone) Width() int {
	return p.Right - p.Left
}

// Height returns the height of the pageZone.
func (p *pageZone) Height() int {
	return p.Bottom - p.Top
}

// page is a web page represented by an image located at URL, holding 0
// or more Zones which, when clicked, redirect to the page in question.
type page struct {
	URL       string     `yaml:"URL"`
	Title     string     `yaml:"Title"`
	ImagePath string     `yaml:"ImagePath"`
	Zones     []pageZone `yaml:"Zones"`
}

// endpoint provides an httphandler for each page.
func (p *page) endpoint(tpl *template.Template) (http.HandlerFunc, error) {
	if _, err := os.Stat(p.ImagePath); err != nil {
		return nil, fmt.Errorf("%s: image %s not found", p.URL, p.ImagePath)
	}
	if len(p.Zones) < 1 {
		return nil, fmt.Errorf("%s: need a least one zone", p.URL)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := tpl.Execute(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}, nil
}
