package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// pageZone sets up a rectangular page zone on a page that, when
// clicked, redirects to target.
type pageZone struct {
	x1, y1 int
	x2, y2 int
	target string
}

// page is an web page represented by an image located at url, holding 0
// or more zones which, when clicked, redirect to the page in question.
type page struct {
	url       string
	title     string
	imagePath string // path
	zones     []pageZone
}

// fileExists reports if a file at path exits.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// endpoint provides an httphandler for each page
func (p *page) endpoint() (http.HandlerFunc, error) {
	if !fileExists(p.imagePath) {
		return nil, fmt.Errorf("%s: image %s not found", p.url, p.imagePath)
	}
	if len(p.zones) < 1 {
		return nil, fmt.Errorf("%s: need a least one zone", p.url)
	}
	zone0 := p.zones[0]
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(
			// rough and ready
			fmt.Sprintf(
				`<html><head>
				       <title>%s</title>
				       <style>
					   .area {
						   z-index: 1;
						   position: absolute;
						   background-color: transparent;
						   left: %dpx;
						   top: %dpx;
						   width: %dpx;
						   height: %dpx;
						}
						.area:hover {
							background-color: blue;
							opacity: 0.2;
						}
						.area:onclick {
							z-index: -1;
						}
					   </style>
				       </head>
				       <body>
					   <img src="%s" usemap="#whatever" />
					   <map name="whatever">
							<area shape="rect" 
							 coords="%d,%d,%d,%d"
							 alt="xxx"
							 href="%s"
					   />
					   <div class="area"></div>
					   </body>
					   </html>`,
				p.title,
				zone0.x1, zone0.y1, zone0.x2-zone0.x1, zone0.y2-zone0.y1,
				p.imagePath,
				zone0.x1, zone0.y1, zone0.x2, zone0.y2,
				zone0.target,
			)))
	}, nil
}

type server struct {
	imageDir      string
	serverAddress string
	serverPort    string
	pages         []page
}

func newServer(pages []page) (*server, error) {
	s := server{
		imageDir:      "./images",
		serverAddress: "127.0.0.1",
		serverPort:    "8000",
	}
	if len(pages) < 1 {
		return nil, fmt.Errorf("at least one page must be provided")
	}
	for _, p := range pages {
		s.pages = append(s.pages, p)
	}
	return &s, nil
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
	// endpoint routing; gorilla mux is used because "/" in http.NewServeMux
	// is a catch-all pattern
	r := mux.NewRouter()

	// atach image serving directory
	// https://eli.thegreenplace.net/2022/serving-static-files-and-web-apps-in-go/
	imgDir := http.FileServer(http.Dir("images"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imgDir))
	// r.Handle("/images/", http.StripPrefix("/images/", imgDir))

	r.HandleFunc("/health", s.Health)
	// r.HandleFunc("/favicon", s.Favicon())

	for _, p := range s.pages {
		pe, err := p.endpoint()
		if err != nil {
			return err
		}
		// add route
		r.HandleFunc(p.url, pe)
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

	// compression handler
	compressor := func(handler http.Handler) http.Handler {
		return handlers.CompressHandler(handler)
	}

	// attach middleware
	// r.Use(bodyLimitMiddleware)
	r.Use(logging)
	r.Use(compressor)
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

func main() {

	pages := []page{
		page{
			url:       "/home",
			title:     "Home",
			imagePath: "images/home.jpg",
			zones: []pageZone{
				pageZone{
					x1:     367,
					y1:     94,
					x2:     539,
					y2:     263,
					target: "/detail",
				},
			},
		},
		page{
			url:       "/detail",
			title:     "Detail",
			imagePath: "images/detail.jpg",
			zones: []pageZone{
				pageZone{
					x1:     436,
					y1:     31,
					x2:     538,
					y2:     73,
					target: "/home",
				},
			},
		},
	}

	server, err := newServer(pages)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = server.serve()
	if err != nil {
		fmt.Println(err)
	}

}
