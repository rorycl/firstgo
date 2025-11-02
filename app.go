package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

// App is the main "plug point" for the application, making the three
// modes of "Serve" (embedded, on disk and development mode) and
// "WriteAssets" injectable into the cli flags package. If the
// interactive flag is set messages are printed to the console.
type App struct {
	interactive bool
	serveFunc   func(*server) error
	writeFunc   func(cfg *config, directory string) error
	stopper     chan struct{} // for tests
}

// NewApp returns a new App.
func NewApp() *App {
	return &App{
		serveFunc: Serve,
		writeFunc: WriteAssets,
	}
}

// Interactive toggles the interactive state. By default this is off.
func (a *App) Interactive() {
	a.interactive = !a.interactive
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
	if a.interactive {
		fmt.Printf("Running server on %s:%s\n", address, port)
		fmt.Printf("(the index is at <http://%s:%s/index>)\n", address, port)
	}
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
	if a.interactive {
		fmt.Printf("Running demo server on %s:%s\n", address, port)
		fmt.Printf("(the index is at <http://%s:%s/index>)\n", address, port)
	}
	return a.serveFunc(server)
}

// Init writes the internal directories and config to disk.
func (a *App) Init(dir string) error {
	config, err := newConfig(configYaml, true) // is bytes
	if err != nil {
		return err
	}
	if a.interactive {
		fmt.Printf("writing demo files to %q\n", dir)
	}
	return a.writeFunc(config, dir)
}

// ServeInDevelopment serves the service from disk in development mode,
// using an extraordinarily elaborate event loop and filesystem watcher
// to reload the configuration and server on changes, waiting for
// further file writes when an error occurs.
func (a *App) ServeInDevelopment(address, port string, templateSuffixes []string, configFile string) error {

	var srv *server
	var cfg *config
	var templateDir = "assets/templates"

	// 1. Define the sets of commands for the event loop.

	// loadConfigCmd is a configuration loader command.
	loadConfigCmd := func(ctx context.Context) Msg {
		configBytes, err := os.ReadFile(configFile)
		if err != nil {
			log.Printf("config file error: %v", err)
			return "FILE_WAIT"
		}

		config, err := newConfig(configBytes, false)
		if err != nil {
			log.Printf("config load error: %v", err)
			log.Println("waiting for file fix")
			return "FILE_WAIT"
		}
		cfg = config
		templateDir = filepath.Join(cfg.AssetsDir, "templates")
		log.Println("config load ok")
		return "CONFIG_LOAD_OK"
	}

	// startServerCmd is a server starting command.
	startServerCmd := func(ctx context.Context) Msg {
		var err error
		if srv != nil {
			_ = srv.webServer.Shutdown(context.Background())
		}
		srv, err = newServer(address, port, cfg)
		if err != nil {
			log.Printf("server start error: %v", err)
			log.Println("waiting for file fix")
			return "FILE_WAIT"
		}
		log.Printf("Running server on %s:%s\n", address, port)
		log.Printf("   (the index is at <http://%s:%s/index>)\n", address, port)

		var wg sync.WaitGroup
		wg.Go(func() {
			// normally a blocking call
			err := a.serveFunc(srv)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("fatal server error: %v", err) // should not happen
			}
		})
		go func() {
			wg.Wait()
		}()
		log.Println("server started ok")
		return "SERVER_STARTED"
	}

	// fileWaitForUpdateCmd is a file watcher command.
	fileWaitForUpdateCmd := func(ctx context.Context) Msg {
		fcn, err := NewFileChangeNotifier(
			[]DirFilesDescriptor{
				DirFilesDescriptor{filepath.Dir(configFile), []string{filepath.Ext(configFile)}},
				DirFilesDescriptor{templateDir, templateSuffixes},
			},
		)
		if err != nil {
			log.Fatalf("error initialising watcher: %v", err)
		}

		watchErrChan := make(chan error)
		go func() {
			watchErrChan <- fcn.Watch(ctx)
		}()

		select {
		case <-watchErrChan:
			log.Printf("file watch error: %v", err)
			log.Println("waiting for file fix")
			return "FILE_WAIT"
		case _, ok := <-fcn.Update():
			if !ok {
				return ""
			}
			log.Println("---------------------------------------")
			log.Println("file update detected")
		}
		return "FILE_UPDATED"
	}

	// 2. make context for the main event loop.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. initiaise and run the event loop.
	el, err := NewEventLoop(
		[]LabelledCmd{
			LabelledCmd{"CONFIG_LOAD_OK", startServerCmd},
			LabelledCmd{"CONFIG_LOAD_FAILED", fileWaitForUpdateCmd},
			LabelledCmd{"FILE_UPDATED", loadConfigCmd},
			LabelledCmd{"SERVER_STARTED", fileWaitForUpdateCmd},
		},
		loadConfigCmd,        // start command
		fileWaitForUpdateCmd, // default command
	)
	if err != nil {
		log.Fatalf("event loop init error: %v", err)
	}

	// app.stopper is for stopping the server in tests.
	if a.stopper != nil {
		go func() {
			<-a.stopper
			log.Println("stopper received")
			cancel()
		}()
	}

	// catch ^C
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		<-c
		log.Println("")
		log.Println("Interrupt received. Shutting down.")
		cancel()
	}()

	// 4. Run the event loop.
	el.Run(ctx)

	return nil
}
