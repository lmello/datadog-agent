// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package inventoryagent

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/comp/core/config"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/pkg/serializer"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func getTestInventoryPayload(t *testing.T, confOverrides map[string]any) *inventoryagent {
	p := newInventoryAgentProvider(
		fxutil.Test[dependencies](
			t,
			log.MockModule,
			config.MockModule,
			fx.Replace(config.MockParams{Overrides: confOverrides}),
			fx.Provide(func() serializer.MetricSerializer { return &serializer.MockSerializer{} }),
		),
	)
	return p.Comp.(*inventoryagent)
}

func TestSet(t *testing.T) {
	ia := getTestInventoryPayload(t, nil)

	ia.Set("test", 1234)
	assert.Equal(t, 1234, ia.data["test"])
}

func TestGetPayload(t *testing.T) {
	ia := getTestInventoryPayload(t, nil)
	ia.hostname = "hostname-for-test"

	ia.Set("test", 1234)
	startTime := time.Now().UnixNano()

	p := ia.getPayload()
	payload := p.(*Payload)

	assert.True(t, payload.Timestamp > startTime)
	assert.Equal(t, "hostname-for-test", payload.Hostname)
	assert.Equal(t, 1234, payload.Metadata["test"])
}
