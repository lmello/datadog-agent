// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package config implements a component to handle agent configuration.  This
// component temporarily wraps pkg/config.
//
// This component initializes pkg/config based on the bundle params, and
// will return the same results as that package.  This is to support migration
// to a component architecture.  When no code still uses pkg/config, that
// package will be removed.
//
// The mock component does nothing at startup, beginning with an empty config.
// It also overwrites the pkg/config.Datadog for the duration of the test.
package config

import (
	"go.uber.org/fx"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
)

// team: agent-shared-components

// LogConfig reads the logger config
type LogConfig config.Reader

// Component is the component type.
type Component interface {
	config.Reader

	// Warnings returns config warnings collected during setup.
	Warnings() *config.Warnings

	// Object returns wrapped config
	Object() config.Reader
}

// Module defines the fx options for this component.
var Module = fxutil.Component(
	fx.Provide(newConfig),
)
