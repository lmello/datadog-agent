// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package log

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"

	"github.com/cihub/seelog"

	"github.com/DataDog/datadog-agent/comp/core/config"
	pkgconfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// logger implements the component
type logger struct {
	// this component is currently implementing a thin wrapper around
	// pkg/util/log, and uses globals in that package.
}

// NewTemporaryLoggerWithoutInit returns a logger component instance. It assumes the logger has already been
// initialized beforehand.
//
// This function should be used when all these conditions are true:
// - You write or update code which uses a lot of logging.
// - You want the code to be components ready.
// - logger.Component cannot be injected.
//
// It should not be used when:
// - You add few logging functions.
// - When the instance of logger.Component is reachable in less than 5 stack frames.
// - It doesn't make the migration to log.Component easier.
func NewTemporaryLoggerWithoutInit() Component {
	return &logger{}
}

func newAgentLogger(lc fx.Lifecycle, params Params, config config.Component) (Component, error) {
	return NewLogger(lc, params, config)
}

// NewLogger creates a log.Component using the provided config.LogConfig
func NewLogger(lc fx.Lifecycle, params Params, config config.LogConfig) (Component, error) {
	if params.logLevelFn == nil {
		return nil, errors.New("must call one of core.BundleParams.ForOneShot or ForDaemon")
	}

	err := pkgconfig.SetupLogger(
		pkgconfig.LoggerName(params.loggerName),
		params.logLevelFn(config),
		params.logFileFn(config),
		params.logSyslogURIFn(config),
		params.logSyslogRFCFn(config),
		params.logToConsoleFn(config),
		params.logFormatJSONFn(config))
	if err != nil {
		return nil, err
	}

	l := &logger{}
	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		l.Flush()
		return nil
	}})

	return l, nil
}

// Until the log migration to component is done, we use *StackDepth to log. The log component add 1 layer to the call
// stack and *StackDepth add another.
//
// We check the current log level to avoid calling Sprintf when it's not needed (Sprintf from Tracef uses a lot a CPU)

// Trace implements Component#Trace.
func (*logger) Trace(v ...interface{}) { log.TraceStackDepth(2, v...) }

// Tracef implements Component#Tracef.
func (*logger) Tracef(format string, params ...interface{}) {
	currentLevel, _ := log.GetLogLevel()
	if currentLevel <= seelog.TraceLvl {
		log.TraceStackDepth(2, fmt.Sprintf(format, params...))
	}
}

// Debug implements Component#Debug.
func (*logger) Debug(v ...interface{}) { log.DebugStackDepth(2, v...) }

// Debugf implements Component#Debugf.
func (*logger) Debugf(format string, params ...interface{}) {
	currentLevel, _ := log.GetLogLevel()
	if currentLevel <= seelog.DebugLvl {
		log.DebugStackDepth(2, fmt.Sprintf(format, params...))
	}
}

// Info implements Component#Info.
func (*logger) Info(v ...interface{}) { log.InfoStackDepth(2, v...) }

// Infof implements Component#Infof.
func (*logger) Infof(format string, params ...interface{}) {
	currentLevel, _ := log.GetLogLevel()
	if currentLevel <= seelog.InfoLvl {
		log.InfoStackDepth(2, fmt.Sprintf(format, params...))
	}
}

// Warn implements Component#Warn.
func (*logger) Warn(v ...interface{}) error { return log.WarnStackDepth(2, v...) }

// Warnf implements Component#Warnf.
func (*logger) Warnf(format string, params ...interface{}) error {
	// no need to check the current log level since Sprintf will be called in all case to generate the returned
	// error
	return log.WarnStackDepth(2, fmt.Sprintf(format, params...))
}

// Error implements Component#Error.
func (*logger) Error(v ...interface{}) error { return log.ErrorStackDepth(2, v...) }

// Errorf implements Component#Errorf.
func (*logger) Errorf(format string, params ...interface{}) error {
	// no need to check the current log level since Sprintf will be called in all case to generate the returned
	// error
	return log.ErrorStackDepth(2, fmt.Sprintf(format, params...))
}

// Critical implements Component#Critical.
func (*logger) Critical(v ...interface{}) error { return log.CriticalStackDepth(2, v...) }

// Criticalf implements Component#Criticalf.
func (*logger) Criticalf(format string, params ...interface{}) error {
	// no need to check the current log level since Sprintf will be called in all case to generate the returned
	// error
	return log.CriticalStackDepth(2, fmt.Sprintf(format, params...))
}

// Flush implements Component#Flush.
func (*logger) Flush() {
	log.Flush()
}
