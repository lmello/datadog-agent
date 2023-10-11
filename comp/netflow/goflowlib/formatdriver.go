// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2022-present Datadog, Inc.

package goflowlib

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-agent/comp/netflow/common"
	flowpb "github.com/netsampler/goflow2/pb"
)

// AggregatorFormatDriver is used as goflow formatter to forward flow data to aggregator/EP Forwarder
type AggregatorFormatDriver struct {
	namespace string
	port      uint16
	flowAggIn chan *common.Flow
	FlowCount map[uint16]int
}

var FlowCountUpdateChan = make(chan common.FlowCountUpdate)

// NewAggregatorFormatDriver returns a new AggregatorFormatDriver
func NewAggregatorFormatDriver(flowAgg chan *common.Flow, namespace string, port uint16, flowCount map[uint16]int) *AggregatorFormatDriver {
	return &AggregatorFormatDriver{
		namespace: namespace,
		port:      port,
		flowAggIn: flowAgg,
		FlowCount: flowCount,
	}
}

// Prepare desc
func (d *AggregatorFormatDriver) Prepare() error {
	return nil
}

// Init desc
func (d *AggregatorFormatDriver) Init(context.Context) error {
	return nil
}

// Format desc
func (d *AggregatorFormatDriver) Format(data interface{}) ([]byte, []byte, error) {
	flow, ok := data.(*flowpb.FlowMessage)
	if !ok {
		return nil, nil, fmt.Errorf("message is not flowpb.FlowMessage")
	}
	d.flowAggIn <- ConvertFlow(flow, d.namespace)
	FlowCountUpdateChan <- common.FlowCountUpdate{Port: int(d.port), Count: 1}
	return nil, nil, nil
}
