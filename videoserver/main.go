package main

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port, found := os.LookupEnv("PORT")
	if !found {
		log.Fatalln("Error: PORT environment variable is not set")
	}

	// Get basic auth credentials from environment
	authUser, found := os.LookupEnv("AUTH_USER")
	if !found {
		log.Fatalln("Error: AUTH_USER environment variable is not set")
	}
	authPass, found := os.LookupEnv("AUTH_PASSWORD")
	if !found {
		log.Fatalln("Error: AUTH_PASSWORD environment variable is not set")
	}

	// Initialize the basic auth middleware
	basicAuth := newBasicAuthMiddleware(authUser, authPass)

	mux := http.NewServeMux()

	// Serve site files from root with basic auth
	siteServer := http.FileServer(http.Dir("/site"))
	mux.Handle("GET /", basicAuth(noCache(siteServer)))

	// Serve archive files with basic auth
	archiveServer := http.FileServer(http.Dir("/archive"))
	mux.Handle("GET /archive/", basicAuth(http.StripPrefix("/archive", archiveServer)))

	// Serve stream files with no-cache and basic auth
	streamServer := http.FileServer(http.Dir("/stream"))
	mux.Handle("GET /stream/", basicAuth(noCache(http.StripPrefix("/stream", streamServer))))

	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s\n", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Error: Unable to start server: %+v\n", err)
	}
}

type noCacheMiddleware struct {
	handler http.Handler
}

func (m noCacheMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// forward the request downstream
	m.handler.ServeHTTP(w, r)
}

func noCache(handler http.Handler) http.Handler {
	return noCacheMiddleware{handler: handler}
}

func newBasicAuthMiddleware(authUser, authPass string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

			if username, password, ok := r.BasicAuth(); !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else {
				// Use constant-time comparison for both username and password
				userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(authUser)) == 1
				passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(authPass)) == 1

				if !userMatch || !passMatch {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			handler.ServeHTTP(w, r)
		})
	}
}
