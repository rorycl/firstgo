package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"html/template"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type WebServer interface {
	ListenAndServe() error
}

const (
	imageDir    = "images"
	staticDir   = "static"
	templateDir = "templates"
)

// server sets the configuration for a simple http server.
type server struct {
	imagePath     string // "/images/"
	staticPath    string
	templatesPath string
	serverAddress string
	serverPort    string
	assetsFS      fs.FS
	pageTpl       *template.Template
	indexTpl      *template.Template
	pages         []page
	indexPages    []string
	webServer     *http.Server
}

// newServer makes a newServer
func newServer(
	address, port string,
	cfg *config,
) (*server, error) {

	if a := net.ParseIP(address); a == nil {
		return nil, fmt.Errorf("invalid IP address: %s", address)
	}
	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	s := server{
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
	s.imagePath = pather(imageDir)
	s.staticPath = pather(staticDir)
	s.templatesPath = pather(templateDir)

	s.assetsFS = cfg.AssetsFS

	var err error

	if len(cfg.Pages) < 2 {
		return nil, errors.New("at least two pages must be provided")
	}
	s.pages = cfg.Pages

	// Attach template.
	s.pageTpl = cfg.PageTpl
	s.indexTpl = cfg.IndexTpl

	// Determine if page indexes are needed.
	s.indexPages = []string{}
	for _, idx := range []string{"/index", "/"} {
		if cfg.hasURL(idx) {
			continue
		}
		s.indexPages = append(s.indexPages, idx)
	}

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
	http.ServeFileFS(w, r, s.assetsFS, "/static/favicon.svg")
}

// Page provides an httphandler for each page.
func (s *server) Page(p *page, tpl *template.Template) (http.HandlerFunc, error) {
	if _, err := fs.Stat(s.assetsFS, p.ImagePath); err != nil {
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

// FourOhFour provides a 404 handler.
func (s *server) FourOhFour(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.Error(w, message, http.StatusNotFound)
	}
}

// Index provides an index of all pages.
func (s *server) Index(pages []page, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := tpl.Execute(w, pages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
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

	// Attach the images and static directories.
	imgFS, err := fs.Sub(s.assetsFS, imageDir)
	if err != nil {
		return nil, fmt.Errorf("image fs mount failure: %w", err)
	}
	r.PathPrefix(s.imagePath).Handler(http.StripPrefix(s.imagePath, http.FileServerFS(imgFS)))

	staticFS, err := fs.Sub(s.assetsFS, staticDir)
	if err != nil {
		return nil, fmt.Errorf("static fs mount failure: %w", err)
	}
	r.PathPrefix(s.staticPath).Handler(http.StripPrefix(s.staticPath, http.FileServerFS(staticFS)))

	// Don't allow /templates to be read
	r.HandleFunc(s.templatesPath, s.FourOhFour(
		"The templates directory is purposely not mounted.",
	))

	r.HandleFunc("/health", s.Health)
	r.HandleFunc("/favicon", s.Favicon)
	r.HandleFunc("/favicon.ico", s.Favicon)

	// Attach the pages defined in the configuration file.
	for _, p := range s.pages {
		pe, err := s.Page(&p, s.pageTpl)
		if err != nil {
			return nil, fmt.Errorf("page build error: %w", err)
		}
		// add route
		r.HandleFunc(p.URL, pe)
	}

	// Attach index pages if required.
	for _, idx := range s.indexPages {
		r.HandleFunc(idx, s.Index(s.pages, s.indexTpl))
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

// Serve starts serving the server at the configured address and port.
func Serve(s *server) error {

	var err error
	s.webServer.Handler, err = s.buildHandler()
	if err != nil {
		return fmt.Errorf("router building error: %w", err)
	}

	err = s.webServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("fatal server error: %w", err)
	}
	return nil
}
