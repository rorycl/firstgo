package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"

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

const (
	AssetDirName   = "assets"
	ConfigFileName = "config.yaml"
)

var RequiredAssetDirs []string = []string{
	"templates",
	"static",
	"images",
}

// Embedded file systems and files
//
//go:embed assets
var assetsFS embed.FS

//go:embed config.yaml
var configYaml []byte

// config describes the config in a yaml configuration file
type config struct {
	PageTemplate  string `yaml:"pageTemplate"`
	IndexTemplate string `yaml:"indexTemplate"`
	Pages         []page `yaml:"pages"`

	// Assets path (for image, template and static directories) and
	// associated fs.FS
	AssetsDir string `yaml:"assetsDir"`
	AssetsFS  fs.FS

	// html templates
	PageTpl  *template.Template
	IndexTpl *template.Template

	pagesByURL   map[string]int
	embeddedMode bool
}

// validateConfig validates the configuration and also sets fields such
// as the filesystems (ImageFS, etc).
func (c *config) validateConfig() error {

	// Attach the filesystems. Beware that embedded filesystems need to
	// be attached below a named container to match the behaviour of
	// os.DirFS.
	if c.embeddedMode {
		var err error
		c.AssetsFS, err = fs.Sub(assetsFS, AssetDirName)
		if err != nil {
			return ErrInvalidConfig{fmt.Sprintf("could not mount embedded fs: %v", err)}
		}
	} else {
		if !dirExists(c.AssetsDir) {
			return ErrInvalidConfig{fmt.Sprintf("directory %q does not exist", c.AssetsDir)}
		}
		c.AssetsFS = os.DirFS(c.AssetsDir)
	}

	// Check the required directories in the AssetsFS
	dir, err := fs.ReadDir(c.AssetsFS, ".")
	if err != nil {
		return fmt.Errorf("internal error: could not read filesystem: %v", err)
	}
OUTER:
	for _, req := range RequiredAssetDirs {
		for _, item := range dir {
			if req == item.Name() && item.Type().IsDir() {
				continue OUTER
			}
		}
		return fmt.Errorf("required directory %q not found in filesystem", req)
	}

	if c.PageTpl, err = template.ParseFS(c.AssetsFS, c.PageTemplate); err != nil {
		return ErrInvalidConfig{fmt.Sprintf("pageTemplate parsing error: %v", err)}
	}
	if c.IndexTpl, err = template.ParseFS(c.AssetsFS, c.IndexTemplate); err != nil {
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
	c.pagesByURL = map[string]int{}

	for ii, pg := range c.Pages {
		if pg.URL == "" {
			return ErrInvalidConfig{fmt.Sprintf("url empty for page %d (%s)", ii, pg.Title)}
		}
		if pg.Title == "" {
			return ErrInvalidConfig{fmt.Sprintf("title empty for page %d (%s)", ii, pg.URL)}
		}
		if pg.ImagePath == "" {
			return ErrInvalidConfig{fmt.Sprintf("image path empty for page %d (%s)", ii, pg.Title)}
		}
		if len(pg.Zones) < 1 {
			return ErrInvalidConfig{fmt.Sprintf("no zones defined for page %d (%s)", ii, pg.Title)}
		}
		if c.hasURL(pg.URL) {
			return ErrInvalidConfig{fmt.Sprintf("URL for page %d (%s) already exists", ii, pg.URL)}
		}
		c.pagesByURL[pg.URL] = ii
	}

	for ii, pg := range c.Pages {
		for zi, zo := range pg.Zones {
			if zo.Target == "" {
				return ErrInvalidConfig{fmt.Sprintf(
					"page %d zone %d empty 'Target' value",
					ii, zi,
				)}
			}
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
			pgIdx, ok := c.pagesByURL[zo.Target]
			if !ok {
				return ErrInvalidConfig{fmt.Sprintf(
					"invalid Zone Target URL %s for page %s (%d) zone %d",
					zo.Target,
					pg.Title,
					ii,
					zi,
				)}
			}
			c.Pages[ii].Zones[zi].TargetTitle = c.Pages[pgIdx].Title
		}
	}
	return nil
}

// hasURL determines if url is in the pages URL field.
func (c *config) hasURL(s string) bool {
	_, ok := c.pagesByURL[s]
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

	TargetTitle string // determined in processing
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
	Note      string     `yaml:"Note",omitempty"`
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

// WriteAssets writes the embedded assets described in the config to
// disk.
func WriteAssets(c *config, savePath string) error {
	if !dirExists(savePath) {
		return fmt.Errorf("directory %s does not exist", savePath)
	}
	if !c.embeddedMode {
		return errors.New("write assets only permitted for embedded mode")
	}

	// Check if the target directory or config files exists
	assetFP := filepath.Join(savePath, AssetDirName)
	if _, err := os.Stat(assetFP); err == nil {
		return fmt.Errorf("target directory %q already exists", assetFP)
	}
	configFP := filepath.Join(savePath, ConfigFileName)
	if _, err := os.Stat(configFP); err == nil {
		return fmt.Errorf("config file %q already exists", configFP)
	}

	// For each embedded FS, write its contents to the corresponding
	// target directory.
	err := writeFSToDisk(assetFP, c.AssetsFS)
	if err != nil {
		return fmt.Errorf("error writing %s: %w", AssetDirName, err)
	}
	return os.WriteFile(configFP, configYaml, 0644)
}

// writeFSToDisk walks an embed.FS and writes its contents to a physical
// directory on disk.
func writeFSToDisk(destRoot string, sourceFS fs.FS) error {
	return fs.WalkDir(sourceFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // propogate errors
		}

		// Destination on disk.
		destPath := filepath.Join(destRoot, path)

		if d.IsDir() {
			// os.MkdirAll is idempotent.
			return os.MkdirAll(destPath, 0755)
		}

		// For files read the content from the virtual file system.
		fileBytes, err := fs.ReadFile(sourceFS, path)
		if err != nil {
			return fmt.Errorf("could not read embedded file %s: %w", path, err)
		}

		// Write the file to the physical disk.
		if err := os.WriteFile(destPath, fileBytes, 0644); err != nil {
			return fmt.Errorf("could not write file to %s: %w", destPath, err)
		}

		return nil
	})
}
