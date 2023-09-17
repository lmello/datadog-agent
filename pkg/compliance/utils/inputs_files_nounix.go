// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build windows

package utils

import "os"

// GetFileUser returns the file user.
func GetFileUser(fi os.FileInfo) string {
	return ""
}

// GetFileGroup returns the file group.
func GetFileGroup(fi os.FileInfo) string {
	return ""
}
