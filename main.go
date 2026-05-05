package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func allowedOrigins() map[string]struct{} {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return nil
	}

	origins := make(map[string]struct{})
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins[origin] = struct{}{}
		}
	}
	return origins
}

func checkOrigin(origins map[string]struct{}) func(*http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		_, ok := origins[origin]
		return ok
	}
}

func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func main() {
	hub := NewHub()
	mux := http.NewServeMux()
	origins := allowedOrigins()
	upgrader := websocket.Upgrader{
		CheckOrigin: checkOrigin(origins),
	}

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade error: %v", err)
			return
		}

		c := &client{
			conn: conn,
			send: make(chan []byte, 64),
		}

		go writePump(c)
		go hub.readPump(c)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := host + ":" + port
	log.Printf("sygy-server listening on %s", addr)
	log.Fatal(newServer(addr, mux).ListenAndServe())
}
