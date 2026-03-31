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

// TestMuxServerV130 verifies that the gorilla/mux v1.3.0 instrumentation
// correctly renames HTTP server spans to the matched route template via the
// BeforeSetCurrentRoute hook (which intercepts the setCurrentRoute function).
func TestMuxServerV130(t *testing.T) {
	testCases := []struct {
		name         string
		port         int
		path         string
		expectedName string
	}{
		{
			// Static route: template == URL path → no rename → span stays "GET"
			name:         "static_route",
			port:         8181,
			path:         "/hello",
			expectedName: "GET",
		},
		{
			// Parametric route: template "/users/{id}" != "/users/42" → rename
			name:         "parametric_route",
			port:         8182,
			path:         "/users/42",
			expectedName: "/users/{id}",
		},
		{
			// Subrouter prefix pattern: "/test/{key}" != "/test/abc" → rename
			name:         "subrouter_prefix_pattern",
			port:         8183,
			path:         "/test/abc",
			expectedName: "/test/{key}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := testutil.NewTestFixture(t)
			f.BuildAndStart("muxserver/v1.3.0", fmt.Sprintf("-port=%d", tc.port))
			testutil.WaitForTCP(t, fmt.Sprintf("127.0.0.1:%d", tc.port))

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d%s", tc.port, tc.path))
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			testutil.WaitForSpanFlush(t)

			span := f.RequireSingleSpan()
			require.Equal(t, tc.expectedName, span.Name())
			testutil.RequireHTTPServerSemconv(
				t,
				span,
				"GET",
				tc.path,
				"http",
				200,
				int64(tc.port),
				"127.0.0.1",
				"Go-http-client/1.1",
				"1.1",
				"127.0.0.1",
			)
		})
	}
}

// TestMuxServerV174 verifies that the gorilla/mux v1.7.4 instrumentation
// correctly renames HTTP server spans to the matched route template via the
// BeforeRequestWithRoute hook (which intercepts the requestWithRoute function,
// available since v1.7.4 with a typed *mux.Route argument).
func TestMuxServerV174(t *testing.T) {
	testCases := []struct {
		name         string
		port         int
		path         string
		expectedName string
	}{
		{
			name:         "static_route",
			port:         8191,
			path:         "/hello",
			expectedName: "GET",
		},
		{
			name:         "parametric_route",
			port:         8192,
			path:         "/users/42",
			expectedName: "/users/{id}",
		},
		{
			// Multi-segment pattern from loongsuite test_mux_prefix.go
			name:         "multi_segment_pattern",
			port:         8193,
			path:         "/france/countries/germany",
			expectedName: "/{name}/countries/{country}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := testutil.NewTestFixture(t)
			f.BuildAndStart("muxserver/v1.7.4", fmt.Sprintf("-port=%d", tc.port))
			testutil.WaitForTCP(t, fmt.Sprintf("127.0.0.1:%d", tc.port))

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d%s", tc.port, tc.path))
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			testutil.WaitForSpanFlush(t)

			span := f.RequireSingleSpan()
			require.Equal(t, tc.expectedName, span.Name())
			testutil.RequireHTTPServerSemconv(
				t,
				span,
				"GET",
				tc.path,
				"http",
				200,
				int64(tc.port),
				"127.0.0.1",
				"Go-http-client/1.1",
				"1.1",
				"127.0.0.1",
			)
		})
	}
}
