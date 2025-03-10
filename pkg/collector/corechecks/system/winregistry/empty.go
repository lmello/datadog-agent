// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !windows

// Package winregistry defines the winregistry check
package winregistry

// Avoid the following error on non-supported platforms:
// "build constraints exclude all Go files in /go/src/github.com/DataDog/datadog-agent/pkg/collector/corechecks/system/winregistry"
func init() {
}
