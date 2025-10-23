package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/goccy/go-yaml"

	"text/template"
)

// ErrInvalidConfig reports an invalid yaml configuration file, although
// one that passed parsing.
type ErrInvalidConfig struct {
	info string
}

// Error reports the error.
func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid yaml configuration: %s", e.info)
}

// config describes the config in a yaml configuration file
type config struct {
	PageTemplate string `yaml:"pageTemplate"`
	Pages        []page `yaml:"pages"`
}

func (c *config) validateConfig() error {
	if c.PageTemplate == "" {
		return ErrInvalidConfig{"pageTemplate not set"}
	}
	if len(c.Pages) < 2 {
		return ErrInvalidConfig{"at least two pages must be defined"}
	}

	// Register of page and zone urls to ensure that the latter only
	// call out valid pages. Note that the context of a page is stored
	// in the zoneURLMap value and if multiple incorrect Zone Target
	// values are used with the same URL the last will be reported.
	urlMap := map[string]bool{}
	zoneURLMap := map[string]string{}

	for ii, pg := range c.Pages {
		if pg.URL == "" {
			return ErrInvalidConfig{fmt.Sprintf("url empty for page %d (%s)", ii, pg.Title)}
		}
		urlMap[pg.URL] = true
		if pg.Title == "" {
			return ErrInvalidConfig{fmt.Sprintf("title empty for page %d (%s)", ii, pg.URL)}
		}
		if pg.ImagePath == "" {
			return ErrInvalidConfig{fmt.Sprintf("image path empty for page %d (%s)", ii, pg.Title)}
		}
		if len(pg.Zones) < 1 {
			return ErrInvalidConfig{fmt.Sprintf("no zones defined for page %d (%s)", ii, pg.Title)}
		}
		for zi, zo := range pg.Zones {
			if zo.Target == "" {
				return ErrInvalidConfig{fmt.Sprintf(
					"page %d zone %d empty 'Target' value",
					ii, zi,
				)}
			}
			zoneURLMap[zo.Target] = fmt.Sprintf(
				"page %d zone %d", ii, zi,
			)
			if zo.Right < zo.Left || zo.Right == 0 {
				return ErrInvalidConfig{fmt.Sprintf(
					"page %d zone %d invalid 'Right' value of %d",
					ii, zi, zo.Right,
				)}
			}
			if zo.Bottom < zo.Top || zo.Bottom == 0 {
				return ErrInvalidConfig{fmt.Sprintf(
					"page %d zone %d invalid 'Bottom' value of %d",
					ii, zi, zo.Bottom,
				)}
			}
		}
	}

	// Validate urls.
	for k, v := range zoneURLMap {
		if _, ok := urlMap[v]; !ok {
			return ErrInvalidConfig{fmt.Sprintf(
				"invalid Zone Target URL %s doesn't point to a page (%s)",
				k, v,
			)}
		}
	}
	return nil
}

// newConfig creates and validates a new config from reading a yaml
// file.
func newConfig(b []byte) (*config, error) {
	var c config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}
	err := c.validateConfig()
	return &c, err
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
