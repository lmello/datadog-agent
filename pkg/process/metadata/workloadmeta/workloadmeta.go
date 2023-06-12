// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package workloadmeta

import (
	"strconv"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/languagedetection"
	"github.com/DataDog/datadog-agent/pkg/languagedetection/languagemodels"
	"github.com/DataDog/datadog-agent/pkg/process/procutil"
	"github.com/DataDog/datadog-agent/pkg/proto/pbgo"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// ProcessEntity is a placeholder workloadmeta.ProcessEntity.
// It does not contain all the fields that the final entity will contain.
type ProcessEntity struct {
	pid      int32
	language *languagemodels.Language
}

// WorkloadMetaExtractor handles enriching processes with languages as well as sending enriched processes to the core agent
// via grpc stream.
type WorkloadMetaExtractor struct {
	// Cache is a map from process hash to the workloadmeta entity
	// The cache key takes the form of `pid:<pid>|createTime:<createTime>`. See hashProcess
	cache      map[string]*ProcessEntity
	cacheMutex sync.RWMutex

	grpcListener mockableGrpcListener
}

// NewWorkloadMetaExtractor constructs the WorkloadMetaExtractor.
func NewWorkloadMetaExtractor() *WorkloadMetaExtractor {
	log.Debug("Instantiated the WorkloadMetaExtractor")
	return &WorkloadMetaExtractor{
		cache:        make(map[string]*ProcessEntity),
		grpcListener: newGrpcListener(),
	}
}

// Extract detects the process language, creates a process entity, and sends that entity to WorkloadMeta
func (w *WorkloadMetaExtractor) Extract(procs map[int32]*procutil.Process) {
	newProcs := make([]*procutil.Process, 0, len(procs))
	newCache := make(map[string]*ProcessEntity, len(procs))
	for pid, proc := range procs {
		hash := hashProcess(pid, proc.Stats.CreateTime)
		if entity, ok := w.cache[hash]; ok {
			newCache[hash] = entity
			continue
		}

		newProcs = append(newProcs, proc)
	}

	newEntities := make([]*ProcessEntity, 0, len(newProcs))
	languages := languagedetection.DetectLanguage(newProcs)
	for i, lang := range languages {
		pid := newProcs[i].Pid
		proc := procs[newProcs[i].Pid]
		entity := &ProcessEntity{
			pid:      pid,
			language: lang,
		}
		newEntities = append(newEntities, entity)
		newCache[hashProcess(pid, proc.Stats.CreateTime)] = entity
	}

	oldProcs := getOldProcs(w.cache, newCache)
	w.grpcListener.writeEvents(oldProcs, newEntities)

	w.cacheMutex.Lock()
	w.cache = newCache
	w.cacheMutex.Unlock()
}

func getOldProcs(oldCache, newCache map[string]*ProcessEntity) []*ProcessEntity {
	oldProcs := make([]*ProcessEntity, 0, len(oldCache))
	for key, entity := range oldCache {
		if _, ok := newCache[key]; ok {
			continue
		}
		oldProcs = append(oldProcs, entity)
	}
	return oldProcs
}

// getCacheState returns the entire state of the cache as a `pbgo.WorkloadMetaProcessEvents`
func (w *WorkloadMetaExtractor) getCacheState() *pbgo.WorkloadMetaProcessEvents {
	w.cacheMutex.RLock()
	defer w.cacheMutex.RUnlock()

	setEvents := make([]*pbgo.ProcessEventSet, 0, len(w.cache))
	for _, proc := range w.cache {
		setEvents = append(setEvents, &pbgo.ProcessEventSet{
			Pid:      proc.pid,
			Language: &pbgo.Language{Name: string(proc.language.Name)},
		})
	}

	return &pbgo.WorkloadMetaProcessEvents{
		SetEvents: setEvents,
	}
}

// Enabled returns whether or not the extractor should be enabled
func Enabled(ddconfig config.ConfigReader) bool {
	return ddconfig.GetBool("process_config.language_detection.enabled")
}

func hashProcess(pid int32, createTime int64) string {
	return "pid:" + strconv.Itoa(int(pid)) + "|createTime:" + strconv.Itoa(int(createTime))
	//return fmt.Sprintf("pid:%v|createTime:%v", pid, createTime)
}