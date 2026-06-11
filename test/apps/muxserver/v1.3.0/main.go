// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package main provides a minimal gorilla/mux HTTP server (v1.3.0) for
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

func main() {
	flag.Parse()

	r := mux.NewRouter()

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

	// Subrouter route (prefix) — used to verify span name = "/test/{key}"
	s := r.PathPrefix("/test").Subrouter()
	s.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("key=" + vars["key"]))
	})

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("starting mux server v1.3.0 on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
