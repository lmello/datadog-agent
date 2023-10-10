// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2022-present Datadog, Inc.

package server

import (
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/comp/netflow/common"
	"github.com/DataDog/datadog-agent/comp/netflow/config"
	"github.com/DataDog/datadog-agent/comp/netflow/flowaggregator"
	"github.com/DataDog/datadog-agent/comp/netflow/goflowlib"
)

// netflowListener contains state of goflow listener and the related netflow config
// flowState can be of type *utils.StateNetFlow/StateSFlow/StateNFLegacy
type netflowListener struct {
	flowState *goflowlib.FlowStateWrapper
	config    config.ListenerConfig
}

// Shutdown will close the goflow listener state
func (l *netflowListener) shutdown() {
	l.flowState.Shutdown()
}

func startFlowListener(listenerConfig config.ListenerConfig, flowAgg *flowaggregator.FlowAggregator, logger log.Component) (*netflowListener, common.FlowListenerData, error) {
	flowState, err := goflowlib.StartFlowRoutine(listenerConfig.FlowType, listenerConfig.BindHost, listenerConfig.Port, listenerConfig.Workers, listenerConfig.Namespace, flowAgg.GetFlowInChan(), logger)
	if err != nil {
		return nil, common.FlowListenerData{}, err
	}

	logger.Debugf("A netflow listener is starting...")

	flowTypeStr := string(listenerConfig.FlowType)

	portTypeInt := int(listenerConfig.Port)

	go func() {
		for update := range goflowlib.FlowCountUpdateChan {
			flowCount := common.GetFlowCountByPort(update.Port)
			flowCount += update.Count
			common.UpdateFlowCountByPort(update.Port, flowCount)
			common.PrintAllFlowDataInstances()
		}
	}()

	flowData := common.FlowListenerData{
		FlowType:  flowTypeStr,
		BindHost:  listenerConfig.BindHost,
		Port:      portTypeInt,
		Workers:   listenerConfig.Workers,
		Namespace: listenerConfig.Namespace,
	}

	common.AddFlowDataInstance(flowData)

	return &netflowListener{
		flowState: flowState,
		config:    listenerConfig,
	}, flowData, nil
}
