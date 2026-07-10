// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package instrument

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otelc/tool/util"
)

func TestToolVersionLine(t *testing.T) {
	marker := "otelc@" + util.Version

	t.Run("release toolchain without rules hash", func(t *testing.T) {
		got := toolVersionLine("compile version go1.26.5", "")
		assert.Equal(t, "compile version go1.26.5 "+marker, got)
	})

	t.Run("release toolchain with rules hash", func(t *testing.T) {
		got := toolVersionLine("compile version go1.26.5", "abcd1234")
		assert.Equal(t, "compile version go1.26.5 "+marker+"/abcd1234", got)
	})

	t.Run("devel toolchain keeps buildID as the last field", func(t *testing.T) {
		line := "compile version devel go1.27-abc123 buildID=x/y/z"
		got := toolVersionLine(line, "abcd1234")
		assert.Equal(t, "compile version devel go1.27-abc123 "+marker+"/abcd1234 buildID=x/y/z", got)
	})
}

func TestLoadMissingMatchedRules(t *testing.T) {
	// Point the work dir at an empty directory: matched.json does not exist,
	// which is what a bare -toolexec build sees when setup never ran.
	t.Setenv(util.EnvOtelcWorkDir, t.TempDir())

	ip := &InstrumentPhase{logger: slog.Default()}
	_, err := ip.load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "otelc setup")
}
