// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package main provides a minimal Elasticsearch client for integration testing.
// Uses the functional (non-typed) client API available in go-elasticsearch v8.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

var addr = flag.String("addr", "http://localhost:9200", "Elasticsearch server address")

func main() {
	flag.Parse()

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{*addr},
	})
	if err != nil {
		log.Fatalf("failed to create elasticsearch client: %v", err)
	}

	// create index
	if res, err := client.Indices.Create("my_index"); err != nil {
		log.Printf("create index failed: %v", err)
	} else {
		res.Body.Close()
	}
	slog.Info("created index", "index", "my_index")

	// index a document
	document := struct {
		Name string `json:"name"`
	}{"go-elasticsearch"}
	data, _ := json.Marshal(document)
	if res, err := client.Index("my_index", bytes.NewReader(data)); err != nil {
		log.Printf("index document failed: %v", err)
	} else {
		res.Body.Close()
	}
	slog.Info("indexed document")

	// get document
	if res, err := client.Get("my_index", "1"); err != nil {
		log.Printf("get document failed: %v", err)
	} else {
		res.Body.Close()
	}
	slog.Info("got document")

	// search documents
	query := `{"query":{"match_all":{}}}`
	if res, err := client.Search(
		client.Search.WithIndex("my_index"),
		client.Search.WithBody(strings.NewReader(query)),
	); err != nil {
		log.Printf("search failed: %v", err)
	} else {
		res.Body.Close()
	}
	slog.Info("searched documents")

	// delete index
	if res, err := client.Indices.Delete([]string{"my_index"}); err != nil {
		log.Printf("delete index failed: %v", err)
	} else {
		res.Body.Close()
	}
	slog.Info("deleted index")
}
