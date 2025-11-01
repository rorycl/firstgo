package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

const testFilePattern = "firstgo_apptest_*"

func writeConfig(t *testing.T, config []byte) string {
	t.Helper()
	tf, err := os.CreateTemp("", testFilePattern)
	if err != nil {
		t.Fatal("create temp error", err)
	}
	err = os.WriteFile(tf.Name(), config, 0600)
	if err != nil {
		t.Fatal("write temp error", err)
	}
	return tf.Name()
}

// makeOKConfig makes a temporary config file on disk or returns it as a
// string.
func makeOKConfig(t *testing.T, asPath bool) string {
	t.Helper()
	config := `
---
assetsDir: "assets"
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
`
	if !asPath {
		return config
	}
	return writeConfig(t, []byte(config))
}

// makeNotOKConfig makes a temporary, broken config file on disk or as a
// string.
func makeNotOKConfig(t *testing.T, asPath bool) string {
	t.Helper()
	config := `
---
assetsDir: "assets"
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
        Target: "/broken"  # a broken target
`
	if !asPath {
		return config
	}
	return writeConfig(t, []byte(config))

}

// make a non existant config (ignore asPath condition, not relevant to
// demo mode).
func makeNotExistentConfig(t *testing.T, asPath bool) string {
	filer := writeConfig(t, []byte(""))
	_ = os.Remove(filer)
	return filer
}

func TestApp(t *testing.T) {

	tests := []struct {
		name        string
		mode        string
		address     string
		app         App
		mkConfig    func(t *testing.T, asPath bool) string
		errContains string
	}{
		{
			name:    "interactive true serve ok",
			mode:    "serve",
			address: "127.0.0.1",
			app: App{
				interactive: true,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig: makeOKConfig,
		},
		{
			name: "interactive false serve ok",
			mode: "serve",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig: makeOKConfig,
			address:  "127.0.0.1",
		},
		{
			name: "interactive false serve fail config failure",
			mode: "serve",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig:    makeNotOKConfig,
			address:     "127.0.0.1",
			errContains: "invalid Zone Target URL",
		},
		{
			name: "interactive false serve fail non existent config",
			mode: "serve",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig:    makeNotExistentConfig,
			address:     "127.0.0.1",
			errContains: "no such file or directory",
		},
		{
			name: "interactive false serve fail address",
			mode: "serve",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig:    makeOKConfig,
			address:     "nonsense",
			errContains: "invalid IP address",
		},
		{
			name: "interactive false serve fail ",
			mode: "serve",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return errors.New("serve fail") },
			},
			mkConfig:    makeOKConfig,
			address:     "127.0.0.1",
			errContains: "serve fail",
		},
		{
			name: "demo ok interactive",
			mode: "demo",
			app: App{
				interactive: true,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig: makeOKConfig,
			address:  "127.0.0.1",
		},
		{
			name: "demo ok",
			mode: "demo",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig: makeOKConfig,
			address:  "127.0.0.1",
		},
		{
			name: "demo serve failure",
			mode: "demo",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return errors.New("demo serve failure") },
			},
			mkConfig:    makeOKConfig,
			address:     "127.0.0.1",
			errContains: "demo serve failure",
		},
		{
			name: "demo fail config failure",
			mode: "demo",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig:    makeNotOKConfig,
			address:     "127.0.0.1",
			errContains: "invalid Zone Target URL",
		},
		{
			name: "demo false serve fail address",
			mode: "demo",
			app: App{
				interactive: false,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig:    makeOKConfig,
			address:     "nonsense",
			errContains: "invalid IP address",
		},
		{
			name: "init ok interactive",
			mode: "init",
			app: App{
				interactive: true,
				writeFunc:   func(cfg *config, directory string) error { return nil },
			},
			mkConfig: makeOKConfig,
		},
		{
			name: "init ok",
			mode: "init",
			app: App{
				interactive: false,
				writeFunc:   func(cfg *config, directory string) error { return nil },
			},
			mkConfig: makeOKConfig,
			address:  "127.0.0.1",
		},
		{
			name: "init failure",
			mode: "init",
			app: App{
				interactive: false,
				writeFunc:   func(cfg *config, directory string) error { return errors.New("init failure") },
			},
			mkConfig:    makeOKConfig,
			errContains: "init failure",
		},
		{
			name: "init fail config failure",
			mode: "init",
			app: App{
				interactive: false,
				writeFunc:   func(cfg *config, directory string) error { return nil },
			},
			mkConfig:    makeNotOKConfig,
			errContains: "invalid Zone Target URL",
		},
		{
			name:    "development server ok",
			mode:    "development",
			address: "127.0.0.1",
			app: App{
				interactive: true,
				serveFunc:   func(*server) error { return nil },
			},
			mkConfig: makeOKConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			switch tt.mode {
			case "serve":
				cleanup := func(fileName string) func() {
					return func() { _ = os.Remove(fileName) }
				}
				config := tt.mkConfig(t, true) // bool is for "asPath" mode
				t.Cleanup(cleanup(config))
				err = tt.app.Serve(tt.address, "8000", config)
			case "demo":
				config := tt.mkConfig(t, false) // config as string only
				orig := configYaml
				configYaml = []byte(config) // override embed
				err = tt.app.Demo(tt.address, "8000")
				configYaml = orig
			case "init":
				config := tt.mkConfig(t, false) // config as string only
				orig := configYaml
				configYaml = []byte(config) // override embed
				err = tt.app.Init("anything goes")
				configYaml = orig
			case "development":
				cleanup := func(fileName string) func() {
					return func() { _ = os.Remove(fileName) }
				}
				config := tt.mkConfig(t, true) // bool is for "asPath" mode
				t.Cleanup(cleanup(config))
				tt.app.stopper = make(chan struct{})
				go func() {
					a := time.After(25 * time.Millisecond)
					<-a
					fmt.Println("stopper fired")
					tt.app.stopper <- struct{}{}
				}()
				err = tt.app.ServeInDevelopment(tt.address, "8000", []string{"html"}, config)
			default:
				t.Fatalf("mode %q not known", tt.mode)
			}
			if err != nil {
				if tt.errContains == "" {
					t.Fatal("unexpected error", err)
				}
				if got, want := err.Error(), tt.errContains; !strings.Contains(got, want) {
					t.Errorf("got err %q want err with %q", got, want)
				}
			}
			if err == nil && tt.errContains != "" {
				t.Fatalf("expected err with %q", tt.errContains)
			}
		})
	}
}

func TestAppNewInteractive(t *testing.T) {
	app := NewApp()
	if app.interactive != false {
		t.Fatal("expected interactive field to be false")
	}
	app.Interactive()
	if app.interactive != true {
		t.Fatal("expected interactive field to be true")
	}
}
