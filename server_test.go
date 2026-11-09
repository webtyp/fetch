//go:build !wasm

package fetch_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// setupTestServer creates a new httptest.Server with predefined handlers for testing.
func setupTestServer() *httptest.Server {
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
		// Strict content-type check relaxed slightly to allow "application/json" without charset
		// or strict match.
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			http.Error(w, "bad content type\n", http.StatusBadRequest)
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

	return httptest.NewServer(mux)
}
