// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux

// Package activitytree holds activitytree related files
package activitytree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNSSplit(t *testing.T) {
	p, s := dnsSplit("1234.datadoghq.com", 1)
	assert.Equal(t, "1234.datadoghq", p)
	assert.Equal(t, "com", s)

	p, s = dnsSplit("1234.datadoghq.com", 2)
	assert.Equal(t, "1234", p)
	assert.Equal(t, "datadoghq.com", s)

	p, s = dnsSplit("1234.datadoghq.com", 3)
	assert.Equal(t, "1234.datadoghq.com", p)
	assert.Equal(t, "", s)

	p, s = dnsSplit("1234.datadoghq.com.", 1)
	assert.Equal(t, "1234.datadoghq.com", p)
	assert.Equal(t, "", s)

	p, s = dnsSplit("1234", 1)
	assert.Equal(t, "1234", p)
	assert.Equal(t, "", s)

	p, s = dnsSplit("cspm-intake.datad0g.com.datadog-agent.svc.parent9.cluster.local", 3)
	assert.Equal(t, "cspm-intake.datad0g.com.datadog-agent.svc", p)
	assert.Equal(t, "parent9.cluster.local", s)
}
