// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"
)

func TestHelloworld(t *testing.T) {
	stdout, stderr := Build(t, "testdata/helloworld")
	Golden(t, FilterJSON(stdout+stderr), "helloworld/expected_output.golden")
}
