// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package main provides a minimal gorilla/mux HTTP server (v1.7.4) for
// integration testing with the otelc compile-time instrumentation tool.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var port = flag.String("port", "8080", "The server port")

// loggingMiddleware is a simple middleware that logs the request URI.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request: %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	// Simple route — used to verify basic span name = "/hello"
	r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	// Parametric route — used to verify span name = "/users/{id}"
	r.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("user=" + vars["id"]))
	})

	// Multi-segment pattern — used to verify span name = "/{name}/countries/{country}"
	r.HandleFunc("/{name}/countries/{country}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(vars["name"] + "/" + vars["country"]))
	}).Methods(http.MethodGet)

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("starting mux server v1.7.4 on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
