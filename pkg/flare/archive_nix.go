// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !windows

package flare

import (
	flaretypes "github.com/DataDog/datadog-agent/comp/core/flare/types"
)

func getWindowsData(fb flaretypes.FlareBuilder) {} //nolint:revive // TODO fix revive unused-parameter
