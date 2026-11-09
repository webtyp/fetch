//go:build !wasm

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Custom")
		w.Header().Set("Access-Control-Expose-Headers", "X-Test-Simple, X-Reflected-X-Custom")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	// Handler for simple GET requests
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get success"))
	})

	// Handler for JSON POST requests, echoes the body back
	mux.HandleFunc("/post_json", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "bad content type", http.StatusBadRequest)
			return
		}
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	// Handler for file uploads
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	mux.HandleFunc("/patch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		b, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("patched " + string(b)))
	})

	mux.HandleFunc("/delete-body", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		b, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("deleted " + string(b)))
	})

	// Handler for PUT requests
	mux.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("put success"))
	})

	// Handler for DELETE requests
	mux.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("delete success"))
	})

	// Handler that reflects headers
	mux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			w.Header().Set("X-Reflected-"+k, strings.Join(v, ","))
		}
		w.Header().Set("X-Test-Simple", "simple value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("headers ok"))
	})

	// Handler that simulates a slow response
	mux.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow response"))
	})

	// Handler that always returns an error status
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	// Shutdown handler
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("shutting down"))
		go func() {
			time.Sleep(100 * time.Millisecond)
			os.Exit(0)
		}()
	})

	// Create listener with dynamic port (0 = let OS choose)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	// Get the actual port assigned
	port := listener.Addr().(*net.TCPAddr).Port
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Write server URL to file for WASM tests (in parent directory)
	if err := os.WriteFile("../.test_server_url", []byte(serverURL), 0644); err != nil {
		log.Printf("Warning: could not write server URL to file: %v", err)
	}

	log.Printf("Test server running on %s", serverURL)
	log.Fatal(http.Serve(listener, corsMiddleware(mux)))
}
