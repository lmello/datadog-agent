// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package run

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/DataDog/datadog-agent/cmd/trace-agent/subcommands"
	coreconfig "github.com/DataDog/datadog-agent/comp/core/config"
	"github.com/DataDog/datadog-agent/comp/trace"
	"github.com/DataDog/datadog-agent/comp/trace/agent"
	"github.com/DataDog/datadog-agent/comp/trace/config"
	pkgconfig "github.com/DataDog/datadog-agent/pkg/config"
	tracelog "github.com/DataDog/datadog-agent/pkg/trace/log"
	"github.com/DataDog/datadog-agent/pkg/trace/telemetry"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// Stack depth of 3 since the `corelogger` struct adds a layer above the logger
const stackDepth = 3

// MakeCommand returns the run subcommand for the 'trace-agent' command.
func MakeCommand(globalParamsGetter func() *subcommands.GlobalParams) *cobra.Command {

	cliParams := &RunParams{}
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Start datadog trace-agent.",
		Long:  `The Datadog trace-agent aggregates, samples, and forwards traces to datadog submitted by tracers loaded into your application.`,
		RunE: func(*cobra.Command, []string) error {
			cliParams.GlobalParams = globalParamsGetter()
			return runTraceAgent(cliParams, cliParams.ConfPath)
		},
	}

	setParamFlags(runCmd, cliParams)

	return runCmd
}

func setParamFlags(cmd *cobra.Command, cliParams *RunParams) {
	cmd.PersistentFlags().StringVarP(&cliParams.PIDFilePath, "pidfile", "p", "", "path for the PID file to be created")
	cmd.PersistentFlags().StringVarP(&cliParams.CPUProfile, "cpu-profile", "l", "",
		"enables CPU profiling and specifies profile path.")
	cmd.PersistentFlags().StringVarP(&cliParams.MemProfile, "mem-profile", "m", "",
		"enables memory profiling and specifies profile.")

	setOSSpecificParamFlags(cmd, cliParams)
}

func runFx(ctx context.Context, cliParams *RunParams, defaultConfPath string) error {
	if cliParams.ConfPath == "" {
		cliParams.ConfPath = defaultConfPath
	}
	err := fxutil.Run(
		// ctx is required to be supplied from here, as Windows needs to inject its own context
		// to allow the agent to work as a service.
		fx.Provide(func() context.Context { return ctx }), // fx.Supply(ctx) fails with a missing type error.
		fx.Supply(coreconfig.NewAgentParamsWithSecrets(cliParams.ConfPath)),
		coreconfig.Module,
		fx.Invoke(func(_ config.Component) {}),
		// Required to avoid cyclic imports.
		fx.Provide(func(cfg config.Component) telemetry.TelemetryCollector { return telemetry.NewCollector(cfg.Object()) }),
		fx.Supply(&agent.Params{
			CPUProfile:  cliParams.CPUProfile,
			MemProfile:  cliParams.MemProfile,
			PIDFilePath: cliParams.PIDFilePath,
		}),
		trace.Bundle,
		// TODO: corelogger must be a component (for future reference)
		fx.Invoke(func(cfg config.Component, telemetryCollector telemetry.TelemetryCollector) error {
			tracecfg := cfg.Object()
			if err := pkgconfig.SetupLogger(
				pkgconfig.LoggerName("TRACE"),
				pkgconfig.Datadog.GetString("log_level"),
				tracecfg.LogFilePath,
				pkgconfig.GetSyslogURI(),
				pkgconfig.Datadog.GetBool("syslog_rfc"),
				pkgconfig.Datadog.GetBool("log_to_console"),
				pkgconfig.Datadog.GetBool("log_format_json"),
			); err != nil {
				telemetryCollector.SendStartupError(telemetry.CantCreateLogger, err)
				return fmt.Errorf("Cannot create logger: %v", err)
			}
			tracelog.SetLogger(corelogger{})
			return nil
		}),
		fx.Invoke(func(_ agent.Component) {}),
	)
	if err != nil && errors.Is(err, agent.ErrAgentDisabled) {
		return nil
	}
	return err
}

type corelogger struct{}

// Trace implements Logger.
func (corelogger) Trace(v ...interface{}) { log.TraceStackDepth(stackDepth, v...) }

// Tracef implements Logger.
func (corelogger) Tracef(format string, params ...interface{}) {
	log.TracefStackDepth(stackDepth, format, params...)
}

// Debug implements Logger.
func (corelogger) Debug(v ...interface{}) { log.DebugStackDepth(stackDepth, v...) }

// Debugf implements Logger.
func (corelogger) Debugf(format string, params ...interface{}) {
	log.DebugfStackDepth(stackDepth, format, params...)
}

// Info implements Logger.
func (corelogger) Info(v ...interface{}) { log.InfoStackDepth(stackDepth, v...) }

// Infof implements Logger.
func (corelogger) Infof(format string, params ...interface{}) {
	log.InfofStackDepth(stackDepth, format, params...)
}

// Warn implements Logger.
func (corelogger) Warn(v ...interface{}) error { return log.WarnStackDepth(stackDepth, v...) }

// Warnf implements Logger.
func (corelogger) Warnf(format string, params ...interface{}) error {
	return log.WarnfStackDepth(stackDepth, format, params...)
}

// Error implements Logger.
func (corelogger) Error(v ...interface{}) error { return log.ErrorStackDepth(stackDepth, v...) }

// Errorf implements Logger.
func (corelogger) Errorf(format string, params ...interface{}) error {
	return log.ErrorfStackDepth(stackDepth, format, params...)
}

// Critical implements Logger.
func (corelogger) Critical(v ...interface{}) error { return log.CriticalStackDepth(stackDepth, v...) }

// Criticalf implements Logger.
func (corelogger) Criticalf(format string, params ...interface{}) error {
	return log.CriticalfStackDepth(stackDepth, format, params...)
}

// Flush implements Logger.
func (corelogger) Flush() { log.Flush() }
