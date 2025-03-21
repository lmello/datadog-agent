// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package fxutil

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"go.uber.org/fx"
)

// Module is a fx.Module for Component with an exported field "Options" to list options
type Module struct {
	fx.Option
	Options []fx.Option
}

// Component is a simple wrapper around fx.Module that automatically determines
// the component name.
func Component(opts ...fx.Option) Module {
	return Module{
		Option:  fx.Module(getComponentName(), opts...),
		Options: opts,
	}
}

// getComponentName gets the component name of the caller's caller.
//
// This must be a package of the form
// `github.com/DataDog/datadog-agent/comp/<bundle>/module`.
func getComponentName() string {
	_, filename, _, ok := runtime.Caller(2)
	if !ok {
		panic("cannot determine component name")
	}
	filename = filepath.ToSlash(filename)
	components := strings.Split(filename, "/")
	if len(components) >= 4 && components[len(components)-4] == "comp" {
		return fmt.Sprintf("comp/%s/%s", components[len(components)-3], components[len(components)-2])
	}
	panic("must be called from a component (comp/<bundle>/<comp>/component.go)")
}

// BundleOptions is a fx.Module for Bundle with an exported field "Options" to list options
type BundleOptions struct {
	fx.Option
	Options []fx.Option
}

// Bundle is a simple wrapper around fx.Module that automatically determines
// the bundle name.
func Bundle(opts ...fx.Option) BundleOptions {
	return BundleOptions{
		Option:  fx.Module(getBundleName(), opts...),
		Options: opts,
	}
}

// getBundleName gets the bundle name of the caller's caller.
//
// This must be a package of the form
// `github.com/DataDog/datadog-agent/comp/<bundle>`.
func getBundleName() string {
	_, filename, _, ok := runtime.Caller(2)
	if !ok {
		panic("cannot determine bundle name")
	}
	filename = filepath.ToSlash(filename)
	components := strings.Split(filename, "/")
	if len(components) >= 3 && components[len(components)-3] == "comp" {
		return fmt.Sprintf("comp/%s", components[len(components)-2])
	}
	panic("must be called from a bundle (comp/<bundle>/bundle.go)")
}
