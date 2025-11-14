// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/ex"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/internal/setup"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/util"
)

//nolint:gochecknoglobals // Implementation of a CLI command
var commandSetup = cli.Command{
	Name:            "setup",
	Description:     "Set up the environment for instrumentation",
	ArgsUsage:       "[go build flags]",
	SkipFlagParsing: true,
	Before:          addLoggerPhaseAttribute,
	Action: func(ctx context.Context, cmd *cli.Command) error {
		logger := util.LoggerFromContext(ctx)
		err := util.BackupFile(backupFiles)
		if err != nil {
			logger.Warn("failed to back up go.mod, go.sum, go.work, go.work.sum, proceeding despite this", "error", err)
		}
		args := cmd.Args().Slice()
		err = setup.Setup(ctx, args, backupFiles)
		if err != nil {
			return ex.Wrapf(err, "failed to setup with exit code %d", exitCodeFailure)
		}
		return nil
	},
}
