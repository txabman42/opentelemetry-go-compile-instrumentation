// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package main provides a minimal Gin HTTP server for integration testing.
// It serves three routes:
//   - GET /user/:name  — parameterised route exercising BeforeNext (c.String)
//   - GET /query       — static route (pattern == raw path, no rename)
//   - GET /tmpl/:id    — parameterised route exercising BeforeHTML (c.HTML)
//
// The server is designed to be instrumented with the otelc compile-time tool
// and validates that gin route patterns are used as span names.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var port = flag.String("port", "8082", "The server port")

func main() {
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Register a minimal inline template so c.HTML() can be called without
	// loading files from disk — keeps the test app self-contained.
	r.SetHTMLTemplate(template.Must(
		template.New("tmpl.html").Parse(`<html><body>id={{.id}}</body></html>`),
	))

	r.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, fmt.Sprintf("hello %s", name))
	})

	r.GET("/query", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/tmpl/:id", func(c *gin.Context) {
		c.HTML(http.StatusOK, "tmpl.html", gin.H{"id": c.Param("id")})
	})

	addr := fmt.Sprintf(":%s", *port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
