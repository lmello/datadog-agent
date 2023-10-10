// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package inventoryagent

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/comp/core/config"
	flaretypes "github.com/DataDog/datadog-agent/comp/core/flare/types"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/comp/metadata/internal/util"
	"github.com/DataDog/datadog-agent/comp/metadata/runner"
	"github.com/DataDog/datadog-agent/pkg/serializer"
	"github.com/DataDog/datadog-agent/pkg/serializer/marshaler"
	"github.com/DataDog/datadog-agent/pkg/util/hostname"
	"go.uber.org/fx"
)

type agentMetadata map[string]interface{}

// Payload handles the JSON unmarshalling of the metadata payload
type Payload struct {
	Hostname  string        `json:"hostname"`
	Timestamp int64         `json:"timestamp"`
	Metadata  agentMetadata `json:"agent_metadata"`
}

// MarshalJSON serialization a Payload to JSON
func (p *Payload) MarshalJSON() ([]byte, error) {
	type PayloadAlias Payload
	return json.Marshal((*PayloadAlias)(p))
}

// SplitPayload implements marshaler.AbstractMarshaler#SplitPayload.
//
// In this case, the payload can't be split any further.
func (p *Payload) SplitPayload(_ int) ([]marshaler.AbstractMarshaler, error) {
	return nil, fmt.Errorf("could not split inventories agent payload any more, payload is too big for intake")
}

type inventoryagent struct {
	util.InventoryPayload

	log      log.Component
	conf     config.Component
	m        sync.Mutex
	data     agentMetadata
	hostname string
}

type dependencies struct {
	fx.In

	Log        log.Component
	Config     config.Component
	Serializer serializer.MetricSerializer
}

type provides struct {
	fx.Out

	Comp          Component
	Provider      runner.Provider
	FlareProvider flaretypes.Provider
}

func newInventoryAgentProvider(deps dependencies) provides {
	hname, _ := hostname.Get(context.Background())
	ia := &inventoryagent{
		conf:     deps.Config,
		log:      deps.Log,
		hostname: hname,
		data:     make(agentMetadata),
	}

	ia.Init(deps.Config, deps.Log, deps.Serializer, ia.getPayload, "agent.json")

	if ia.Enabled {
		ia.initData()
	}

	return provides{
		Comp:          ia,
		Provider:      ia.MetadataProvider(),
		FlareProvider: ia.FlareProvider(),
	}
}

// Set updates a metadata value in the payload. The given value will be stored in the cache without being copied. It is
// up to the caller to make sure the given value will not be modified later.
func (ia *inventoryagent) Set(name string, value interface{}) {
	if !ia.Enabled {
		return
	}

	ia.log.Debugf("setting inventory agent metadata '%s': '%v'", name, value)

	ia.m.Lock()
	defer ia.m.Unlock()

	if !reflect.DeepEqual(ia.data[name], value) {
		ia.data[name] = value
		ia.Refresh()
	}
}

func (ia *inventoryagent) getPayload() marshaler.JSONMarshaler {
	// Create a static copy of agentMetadata for the payload
	data := make(agentMetadata)
	for k, v := range ia.data {
		data[k] = v
	}

	if fullConf, err := ia.getFullAgentConfiguration(); err == nil {
		data["full_configuration"] = fullConf
	}
	if providedConf, err := ia.getProvidedAgentConfiguration(); err == nil {
		data["provided_configuration"] = providedConf
	}

	return &Payload{
		Hostname:  ia.hostname,
		Timestamp: time.Now().UnixNano(),
		Metadata:  data,
	}
}
