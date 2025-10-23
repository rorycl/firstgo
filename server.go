package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"html/template"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type WebServer interface {
	ListenAndServe() error
}

// server sets the configuration for a simple http server.
type server struct {
	imageDir      string // "./images"
	imagePath     string // "/images/"
	staticDir     string
	staticPath    string
	serverAddress string
	serverPort    string
	pages         []page
	htmlTemplate  *template.Template
	webServer     *http.Server
}

// newServer makes a newServer
func newServer(
	address, port string,
	pages []page,
	htmlTemplatePath string,
) (*server, error) {

	s := server{
		imageDir:      "images",
		staticDir:     "static",
		serverAddress: address,
		serverPort:    port,
	}

	// The default server is an http.Server. This can be overridden for
	// testing.
	s.webServer = &http.Server{
		Addr: s.serverAddress + ":" + s.serverPort,
		// timeouts and limits
		// MaxHeaderBytes:    s.WebMaxHeaderBytes,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	pather := func(dir string) string {
		return "/" + filepath.Base(dir) + "/"
	}
	s.imagePath = pather(s.imageDir)
	s.staticPath = pather(s.staticDir)

	var err error
	s.htmlTemplate, err = template.ParseFiles(htmlTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("template parsing error: %v", err)
	}

	if len(pages) < 1 {
		return nil, errors.New("at least one page must be provided")
	}
	s.pages = append(s.pages, pages...)
	return &s, err
}

// HealthCheck shows if the service is up
func (s *server) Health(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := map[string]string{"status": "up"}
	if err := enc.Encode(resp); err != nil {
		log.Print("health error: unable to encode response")
	}
}

// Favicon serves up the favicon
func (s *server) Favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join(s.staticDir, "/favicon.svg"))
}

// buildHandler builds the http handler.
//
// In addition to the pages provided in the pages configuration, a
// "health" and "favicon" endpoint are provided, the first for
// deployment purposes.
func (s *server) buildHandler() (http.Handler, error) {

	// Endpoint routing; gorilla mux is used because "/" in http.NewServeMux
	// is a catch-all pattern.
	r := mux.NewRouter()

	// Atach image and static serving directories.
	imgDir := http.FileServer(http.Dir(s.imageDir))
	r.PathPrefix(s.imagePath).Handler(http.StripPrefix(s.imagePath, imgDir))

	staticDir := http.FileServer(http.Dir(s.staticDir))
	r.PathPrefix(s.staticPath).Handler(http.StripPrefix(s.staticPath, staticDir))

	r.HandleFunc("/health", s.Health)
	r.HandleFunc("/favicon", s.Favicon)

	// Attach the pages defined in the configuration file.
	for _, p := range s.pages {
		pe, err := p.endpoint(s.htmlTemplate)
		if err != nil {
			return nil, err
		}
		// add route
		r.HandleFunc(p.URL, pe)
	}

	// logging converts gorilla's handlers.CombinedLoggingHandler to a
	// func(http.Handler) http.Handler to satisfy type MiddlewareFunc
	logging := func(handler http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(os.Stdout, handler)
	}

	// recovery converts gorilla's handlers.RecoveryHandler to a
	// func(http.Handler) http.Handler to satisfy type MiddlewareFunc
	recovery := func(handler http.Handler) http.Handler {
		return handlers.RecoveryHandler()(handler)
	}

	// attach middleware
	r.Use(logging)
	r.Use(recovery)

	return r, nil
}

// serve starts serving the server at the configured address and port.
func (s *server) serve() error {

	var err error
	s.webServer.Handler, err = s.buildHandler()
	if err != nil {
		return fmt.Errorf("router building error: %w", err)
	}

	err = s.webServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("fatal server error: %v", err)
	}
	return nil
}
