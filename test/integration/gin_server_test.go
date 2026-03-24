// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/test/testutil"
)

const ginServerPort = 8082

// TestGinServer verifies that the gin instrumentation renames HTTP server spans
// to the matched route pattern rather than the raw request URL path.
func TestGinServer(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedSpan string
	}{
		{
			// BeforeNext path: for parameterised routes, gin renames the nethttp
			// server span to the route pattern so telemetry groups by route
			// rather than by raw URL path.
			name:         "BeforeNext — parameterised route",
			path:         "/user/alice",
			expectedSpan: "/user/:name",
		},
		{
			// BeforeHTML path: verifies that c.HTML() on a parameterised route
			// also triggers the span rename via the BeforeHTML hook.
			name:         "BeforeHTML — parameterised route",
			path:         "/tmpl/42",
			expectedSpan: "/tmpl/:id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := testutil.NewTestFixture(t)

			f.BuildAndStart("ginserver", fmt.Sprintf("-port=%d", ginServerPort))
			testutil.WaitForTCP(t, fmt.Sprintf("127.0.0.1:%d", ginServerPort))

			url := fmt.Sprintf("http://127.0.0.1:%d%s", ginServerPort, tc.path)
			resp, err := http.Get(url)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
			testutil.WaitForSpanFlush(t)

			// The gin hook renames the nethttp server span to the route pattern.
			span := testutil.RequireSpan(t, f.Traces(),
				testutil.IsServer,
				testutil.HasAttribute("url.path", tc.path),
			)
			require.Equal(t, tc.expectedSpan, span.Name(),
				"gin instrumentation should rename the span to the route pattern")
			testutil.RequireHTTPServerSemconv(
				t,
				span,
				"GET",
				tc.path,
				"http",
				200,
				int64(ginServerPort),
				"127.0.0.1",
				"Go-http-client/1.1",
				"1.1",
				"127.0.0.1",
			)
		})
	}
}
