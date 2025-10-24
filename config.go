package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
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

// Embedded file systems and files
//
//go:embed images
var imageFS embed.FS

//go:embed templates
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed config.yaml
var configYaml []byte

// config describes the config in a yaml configuration file
type config struct {
	PageTemplate  string `yaml:"pageTemplate"`
	IndexTemplate string `yaml:"indexTemplate"`
	Pages         []page `yaml:"pages"`

	// image, template and static directory paths.
	ImageDir    string `yaml:"imageDir"`
	TemplateDir string `yaml:"templateDir"`
	StaticDir   string `yaml:"staticDir"`

	// image, template and static directories filesystems, either from
	// the embedded assets noted above, or as defined in the config
	// file.
	ImageFS    fs.FS
	TemplateFS fs.FS
	StaticFS   fs.FS

	// html templates
	PageTpl  *template.Template
	IndexTpl *template.Template

	embeddedMode bool
	urlMap       map[string]bool
}

// validateConfig validates the configuration and also sets fields such
// as the filesystems (ImageFS, etc).
func (c *config) validateConfig() error {

	// Attach the filesystems. Beware that embedded filesystems need to
	// be attached below a named container to match the behaviour of
	// os.DirFS.
	if c.embeddedMode {
		var err error
		c.ImageFS, err = fs.Sub(imageFS, "images")
		if err != nil {
			return ErrInvalidConfig{fmt.Sprintf("directory %q could not be mounted", "images")}
		}
		c.TemplateFS, err = fs.Sub(templateFS, "templates")
		if err != nil {
			return ErrInvalidConfig{fmt.Sprintf("directory %q could not be mounted", "templates")}
		}
		c.StaticFS, err = fs.Sub(staticFS, "static")
		if err != nil {
			return ErrInvalidConfig{fmt.Sprintf("directory %q could not be mounted", "static")}
		}
	} else {
		if !dirExists(c.ImageDir) {
			return ErrInvalidConfig{fmt.Sprintf("directory %q does not exist", c.ImageDir)}
		}
		c.ImageFS = os.DirFS(c.ImageDir)

		if !dirExists(c.TemplateDir) {
			return ErrInvalidConfig{fmt.Sprintf("directory %q does not exist", c.TemplateDir)}
		}
		c.TemplateFS = os.DirFS(c.TemplateDir)

		if !dirExists(c.StaticDir) {
			return ErrInvalidConfig{fmt.Sprintf("directory %q does not exist", c.StaticDir)}
		}
		c.StaticFS = os.DirFS(c.StaticDir)
	}

	// Check the template files.
	var err error
	if c.PageTpl, err = template.ParseFS(c.TemplateFS, c.PageTemplate); err != nil {
		return ErrInvalidConfig{fmt.Sprintf("pageTemplate parsing error: %v", err)}
	}
	if c.IndexTpl, err = template.ParseFS(c.TemplateFS, c.IndexTemplate); err != nil {
		return ErrInvalidConfig{fmt.Sprintf("indexTemplate parsing error: %v", err)}
	}

	// Ensure at least two pages are defined.
	if len(c.Pages) < 2 {
		return ErrInvalidConfig{"at least two pages must be defined"}
	}

	// Register of page and zone urls to ensure that the latter only
	// call out valid pages. Note that the context of a page is stored
	// in the zoneURLMap value and if multiple incorrect Zone Target
	// values are used with the same URL the last will be reported.
	c.urlMap = map[string]bool{}
	zoneURLMap := map[string]string{}

	for ii, pg := range c.Pages {
		if pg.URL == "" {
			return ErrInvalidConfig{fmt.Sprintf("url empty for page %d (%s)", ii, pg.Title)}
		}
		if c.hasURL(pg.URL) {
			fmt.Printf("page %s urls %#v\n", pg.URL, c.urlMap)
			return ErrInvalidConfig{fmt.Sprintf("URL for page %d (%s) already exists", ii, pg.URL)}
		}
		c.urlMap[pg.URL] = true
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
		if _, ok := c.urlMap[k]; !ok {
			pageURLS := []string{}
			for p := range c.urlMap {
				pageURLS = append(pageURLS, p)
			}
			return ErrInvalidConfig{fmt.Sprintf(
				"invalid Zone Target URL %s doesn't point to a page (%s)\npages: %s",
				k, v, strings.Join(pageURLS, ", "),
			)}
		}
	}
	return nil
}

// WriteAssets writes the assets to disk.
func (c *config) WriteAssets(savePath string) error {
	if !dirExists(savePath) {
		return fmt.Errorf("directory %s does not exist", savePath)
	}
	if !c.embeddedMode {
		return errors.New("write assets only permitted for embedded mode")
	}
	dirs := map[string]fs.FS{
		"images":    c.ImageFS,
		"templates": c.TemplateFS,
		"static":    c.StaticFS,
	}
	config := "config.yaml"

	// check if any of the dirs or config already exist
	for d := range dirs {
		fp := filepath.Join(savePath, d)
		if dirExists(fp) {
			return fmt.Errorf("filepath %s already exists", fp)
		}
	}
	fp := filepath.Join(savePath, config)
	var pe *os.PathError
	if _, err := os.Stat(fp); !errors.As(err, &pe) {
		return fmt.Errorf("config file %s already exists", fp)
	}

	// for each dir and config, write to path
	descentWriter := func(path string, dfs fs.FS) error {
		err := os.Mkdir(path, 0755)
		if err != nil {
			return fmt.Errorf("could not make dir %s: %v", path, err)
		}
		// always start at the root of the fs
		return fs.WalkDir(dfs, ".", func(iPath string, d fs.DirEntry, err error) error {
			fmt.Printf("%s %v (%T) err %v\n", path, d, d, err)
			if d.IsDir() {
				fp := filepath.Join(path, iPath)
				err := os.Mkdir(fp, 0755)
				if err != nil {
					return fmt.Errorf("could not make dir %s: %v", fp, err)
				}
				return nil
			}
			fp := filepath.Join(path, iPath)
			b, err := fs.ReadFile(dfs, fp)
			if err != nil {
				return fmt.Errorf("could not read file %s: %v", fp, err)
			}
			err = os.WriteFile(fp, b, 0644)
			if err != nil {
				return fmt.Errorf("could not write file %s: %v", fp, err)
			}
			return nil
		})
	}
	for dir, dfs := range dirs {
		err := descentWriter(filepath.Join(savePath, dir), dfs)
		if err != nil {
			fmt.Errorf("writing error: %w", err)
		}
	}
	err := os.WriteFile(config, configYaml, 0644)
	return err
}

// hasURL determines if url is in the pages URL field.
func (c *config) hasURL(s string) bool {
	_, ok := c.urlMap[s]
	return ok
}

// newConfig creates and validates a new config from reading a yaml
// file, initialising in embedded mode or not.
func newConfig(b []byte, embeddedMode bool) (*config, error) {
	var c config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}
	c.embeddedMode = embeddedMode
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

// dirExists checks if the path is to a valid directory.
func dirExists(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !s.IsDir() {
		return false
	}
	return true
}
