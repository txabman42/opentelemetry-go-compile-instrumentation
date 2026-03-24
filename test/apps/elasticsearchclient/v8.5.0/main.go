// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package main provides a minimal Elasticsearch typed client for integration testing.
// Uses the typed (strongly-typed) client API available since go-elasticsearch v8.5.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"log/slog"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

var addr = flag.String("addr", "http://localhost:9200", "Elasticsearch server address")

func main() {
	flag.Parse()

	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{*addr},
	})
	if err != nil {
		log.Fatalf("failed to create elasticsearch typed client: %v", err)
	}

	ctx := context.Background()

	// create index
	if _, err := client.Indices.Create("my_index").Do(ctx); err != nil {
		log.Printf("create index failed: %v", err)
	}
	slog.Info("created index", "index", "my_index")

	// index a document
	document := struct {
		Name string `json:"name"`
	}{"go-elasticsearch"}
	if _, err := client.Index("my_index").Id("1").Request(document).Do(ctx); err != nil {
		log.Printf("index document failed: %v", err)
	}
	slog.Info("indexed document")

	// get document
	if _, err := client.Get("my_index", "1").Do(ctx); err != nil {
		log.Printf("get document failed: %v", err)
	}
	slog.Info("got document")

	// search documents
	if _, err := client.Search().Index("my_index").
		Request(&search.Request{Query: &types.Query{MatchAll: &types.MatchAllQuery{}}}).
		Do(ctx); err != nil {
		log.Printf("search failed: %v", err)
	}
	slog.Info("searched documents")

	// update document
	if _, err := client.Update("my_index", "1").
		Request(&update.Request{Doc: json.RawMessage(`{"name":"updated"}`)}).
		Do(ctx); err != nil {
		log.Printf("update document failed: %v", err)
	}
	slog.Info("updated document")

	// delete document
	if _, err := client.Delete("my_index", "1").Do(ctx); err != nil {
		log.Printf("delete document failed: %v", err)
	}
	slog.Info("deleted document")

	// delete index
	if _, err := client.Indices.Delete("my_index").Do(ctx); err != nil {
		log.Printf("delete index failed: %v", err)
	}
	slog.Info("deleted index")
}
