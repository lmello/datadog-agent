// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package workloadmeta

import (
	"context"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/comp/core/config"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"go.uber.org/fx"
)

// store is a central storage of metadata about workloads. A workload is any
// unit of work being done by a piece of software, like a process, a container,
// a kubernetes pod, or a task in any cloud provider.
type workloadmeta struct {
	log    log.Component
	config config.Component

	// Store related
	storeMut sync.RWMutex
	store    map[Kind]map[string]*cachedEntity // store[entity.Kind][entity.ID] = &cachedEntity{}

	subscribersMut sync.RWMutex
	subscribers    []subscriber

	collectorMut sync.RWMutex
	candidates   map[string]Collector
	collectors   map[string]Collector

	eventCh chan []CollectorEvent

	ongoingPullsMut sync.Mutex
	ongoingPulls    map[string]time.Time // collector ID => time when last pull started
}

type dependencies struct {
	fx.In

	Log     log.Component
	Config  config.Component
	Catalog CollectorList
}

func newWorkloadMeta(lc fx.Lifecycle, ctx context.Context, deps dependencies) Component {

	candidates := make(map[string]Collector)
	for _, c := range deps.Catalog {
		candidates[c.GetID()] = c
	}

	wm := &workloadmeta{
		store:        make(map[Kind]map[string]*cachedEntity),
		candidates:   candidates,
		collectors:   make(map[string]Collector),
		eventCh:      make(chan []CollectorEvent, eventChBufferSize),
		ongoingPulls: make(map[string]time.Time),
	}

	// TODO: we probably need something here
	lc.Append(fx.Hook{OnStart: func(context.Context) error {
		wm.Start(ctx)
		return nil
	}})
	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		return nil
	}})

	return wm
}