package main

import (
	"fmt"
	"os"
)

// App is the main "plug point" for the application, making the two
// modes of "Serve" (embedded and on disk) and "WriteAssets" injectable
// into the cli flags package.
type App struct {
	serveFunc func(*server) error
	writeFunc func(cfg *config, directory string) error
}

// NewApp returns a new App.
func NewApp() *App {
	return &App{
		serveFunc: Serve,
		writeFunc: WriteAssets,
	}
}

// Serve serves the service from disk.
func (a *App) Serve(address, port, configFile string) error {
	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	config, err := newConfig(configBytes, false)
	if err != nil {
		return err
	}

	server, err := newServer(address, port, config)
	if err != nil {
		return err
	}
	fmt.Printf("running server on %s:%s\n", address, port)
	fmt.Printf("the index is at http://%s/index\n", address)
	return a.serveFunc(server)
}

// Demo serves the service from embedded assets.
func (a *App) Demo(address, port string) error {
	config, err := newConfig(configYaml, true) // is bytes
	if err != nil {
		return err
	}

	server, err := newServer(address, port, config)
	if err != nil {
		return err
	}
	fmt.Printf("running demo server on %s:%s\n", address, port)
	fmt.Printf("the index is at http://%s/index\n", address)
	return a.serveFunc(server)
}

// Init writes the internal directories and config to disk.
func (a *App) Init(dir string) error {
	config, err := newConfig(configYaml, true) // is bytes
	if err != nil {
		return err
	}
	fmt.Printf("writing demo files to %q\n", dir)
	return a.writeFunc(config, dir)
}
