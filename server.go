package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"text/template"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// server is a basic http server.
type server struct {
	imageDir      string
	staticDir     string
	serverAddress string
	serverPort    string
	pages         []page
	htmlTemplate  *template.Template
}

// newServer makes a newServer
func newServer(pages []page, htmlTemplatePath string) (*server, error) {
	s := server{
		imageDir:      "./images",
		staticDir:     "./static",
		serverAddress: "127.0.0.1",
		serverPort:    "8000",
	}
	var err error
	s.htmlTemplate, err = template.ParseFiles(htmlTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("template parsing error: %v", err)
	}

	if len(pages) < 1 {
		return nil, errors.New("at least one page must be provided")
	}
	for _, p := range pages {
		s.pages = append(s.pages, p)
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
// func (s *server) Favicon(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFileFS(w, r, s.DirFS.StaticFS, "/favicon.svg")
// }

func (s *server) serve() error {
	// Endpoint routing; gorilla mux is used because "/" in http.NewServeMux
	// is a catch-all pattern.
	r := mux.NewRouter()

	// Atach image and static serving directories.
	// https://eli.thegreenplace.net/2022/serving-static-files-and-web-apps-in-go/
	// Note that paths work differently between the standard http handle
	// and gorilla's mux; PathPrefix is needed in the latter case.
	imgDir := http.FileServer(http.Dir("images"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imgDir))
	staticDir := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticDir))

	r.HandleFunc("/health", s.Health)
	// r.HandleFunc("/favicon", s.Favicon())

	for _, p := range s.pages {
		pe, err := p.endpoint(s.htmlTemplate)
		if err != nil {
			return err
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

	// configure server options
	server := &http.Server{
		Addr:    s.serverAddress + ":" + s.serverPort,
		Handler: r,
		// timeouts and limits
		// MaxHeaderBytes:    s.WebMaxHeaderBytes,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	log.Printf("serving on %s:%s", s.serverAddress, s.serverPort)

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("fatal server error: %v", err)
	}
	return nil
}
