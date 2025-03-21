// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package encoding

import (
	"bytes"
	"math"
	"reflect"
	"unsafe"

	"github.com/twmb/murmur3"

	model "github.com/DataDog/agent-payload/v5/process"
	"github.com/DataDog/datadog-agent/pkg/network"
	"github.com/DataDog/datadog-agent/pkg/process/util"
)

const maxRoutes = math.MaxInt32

// RouteIdx stores the route and the index into the route collection for a route
type RouteIdx struct {
	Idx   int32
	Route model.Route
}

type ipCache map[util.Address]string

func (ipc ipCache) Get(addr util.Address) string {
	if v, ok := ipc[addr]; ok {
		return v
	}

	v := addr.String()
	ipc[addr] = v
	return v
}

// FormatConnection converts a ConnectionStats into an model.Connection
func FormatConnection(builder *model.ConnectionBuilder, conn network.ConnectionStats, routes map[string]RouteIdx, httpEncoder *httpEncoder, http2Encoder *http2Encoder, kafkaEncoder *kafkaEncoder, dnsFormatter *dnsFormatter, ipc ipCache, tagsSet *network.TagsSet) {

	builder.SetPid(int32(conn.Pid))

	var containerID string
	if conn.ContainerID != nil {
		containerID = *conn.ContainerID
	}

	builder.SetLaddr(func(w *model.AddrBuilder) {
		w.SetIp(ipc.Get(conn.Source))
		w.SetPort(int32(conn.SPort))
		w.SetContainerId(containerID)
	})
	builder.SetRaddr(func(w *model.AddrBuilder) {
		w.SetIp(ipc.Get(conn.Dest))
		w.SetPort(int32(conn.DPort))
	})
	builder.SetFamily(uint64(formatFamily(conn.Family)))
	builder.SetType(uint64(formatType(conn.Type)))
	builder.SetIsLocalPortEphemeral(uint64(formatEphemeralType(conn.SPortIsEphemeral)))
	builder.SetLastBytesSent(conn.Last.SentBytes)
	builder.SetLastBytesReceived(conn.Last.RecvBytes)
	builder.SetLastPacketsSent(conn.Last.SentPackets)
	builder.SetLastRetransmits(conn.Last.Retransmits)
	builder.SetLastPacketsReceived(conn.Last.RecvPackets)
	builder.SetDirection(uint64(formatDirection(conn.Direction)))
	builder.SetNetNS(conn.NetNS)
	if conn.IPTranslation != nil {
		builder.SetIpTranslation(func(w *model.IPTranslationBuilder) {
			ipt := formatIPTranslation(conn.IPTranslation, ipc)
			w.SetReplSrcPort(ipt.ReplSrcPort)
			w.SetReplDstPort(ipt.ReplDstPort)
			w.SetReplSrcIP(ipt.ReplSrcIP)
			w.SetReplDstIP(ipt.ReplDstIP)
		})
	}
	builder.SetRtt(conn.RTT)
	builder.SetRttVar(conn.RTTVar)
	builder.SetIntraHost(conn.IntraHost)
	builder.SetLastTcpEstablished(conn.Last.TCPEstablished)
	builder.SetLastTcpClosed(conn.Last.TCPClosed)
	builder.SetProtocol(func(w *model.ProtocolStackBuilder) {
		ps := formatProtocolStack(conn.ProtocolStack, conn.StaticTags)
		for _, p := range ps.Stack {
			w.AddStack(uint64(p))
		}
	})

	builder.SetRouteIdx(formatRouteIdx(conn.Via, routes))
	dnsFormatter.FormatConnectionDNS(conn, builder)

	var (
		staticTags  uint64
		dynamicTags map[string]struct{}
	)
	staticTags, dynamicTags = httpEncoder.GetHTTPAggregationsAndTags(conn, builder)
	_, _ = http2Encoder.WriteHTTP2AggregationsAndTags(conn, builder)

	// TODO: optimize kafkEncoder to take a writer and use gostreamer
	if dsa := kafkaEncoder.GetKafkaAggregations(conn); dsa != nil {
		builder.SetDataStreamsAggregations(func(b *bytes.Buffer) {
			b.Write(dsa)
		})
	}

	conn.StaticTags |= staticTags
	tags, tagChecksum := formatTags(conn, tagsSet, dynamicTags)
	for _, t := range tags {
		builder.AddTags(t)
	}
	builder.SetTagsChecksum(tagChecksum)
}

// FormatCompilationTelemetry converts telemetry from its internal representation to a protobuf message
func FormatCompilationTelemetry(builder *model.ConnectionsBuilder, telByAsset map[string]network.RuntimeCompilationTelemetry) {
	if telByAsset == nil {
		return
	}

	for asset, tel := range telByAsset {
		builder.AddCompilationTelemetryByAsset(func(kv *model.Connections_CompilationTelemetryByAssetEntryBuilder) {
			kv.SetKey(asset)
			kv.SetValue(func(w *model.RuntimeCompilationTelemetryBuilder) {
				w.SetRuntimeCompilationEnabled(tel.RuntimeCompilationEnabled)
				w.SetRuntimeCompilationResult(uint64(tel.RuntimeCompilationResult))
				w.SetRuntimeCompilationDuration(tel.RuntimeCompilationDuration)
			})
		})
	}
}

// FormatConnectionTelemetry converts telemetry from its internal representation to a protobuf message
func FormatConnectionTelemetry(builder *model.ConnectionsBuilder, tel map[network.ConnTelemetryType]int64) {
	// Fetch USM payload telemetry
	ret := GetUSMPayloadTelemetry()

	// Merge it with NPM telemetry
	for k, v := range tel {
		ret[string(k)] = v
	}

	for k, v := range ret {
		builder.AddConnTelemetryMap(func(w *model.Connections_ConnTelemetryMapEntryBuilder) {
			w.SetKey(k)
			w.SetValue(v)
		})
	}

}

// FormatCORETelemetry writes the CORETelemetryByAsset map into a connections payload
func FormatCORETelemetry(builder *model.ConnectionsBuilder, telByAsset map[string]int32) {
	if telByAsset == nil {
		return
	}

	for asset, tel := range telByAsset {
		builder.AddCORETelemetryByAsset(func(w *model.Connections_CORETelemetryByAssetEntryBuilder) {
			w.SetKey(asset)
			w.SetValue(uint64(tel))
		})
	}
}

func formatFamily(f network.ConnectionFamily) model.ConnectionFamily {
	switch f {
	case network.AFINET:
		return model.ConnectionFamily_v4
	case network.AFINET6:
		return model.ConnectionFamily_v6
	default:
		return -1
	}
}

func formatType(f network.ConnectionType) model.ConnectionType {
	switch f {
	case network.TCP:
		return model.ConnectionType_tcp
	case network.UDP:
		return model.ConnectionType_udp
	default:
		return -1
	}
}

func formatDirection(d network.ConnectionDirection) model.ConnectionDirection {
	switch d {
	case network.INCOMING:
		return model.ConnectionDirection_incoming
	case network.OUTGOING:
		return model.ConnectionDirection_outgoing
	case network.LOCAL:
		return model.ConnectionDirection_local
	case network.NONE:
		return model.ConnectionDirection_none
	default:
		return model.ConnectionDirection_unspecified
	}
}

func formatEphemeralType(e network.EphemeralPortType) model.EphemeralPortState {
	switch e {
	case network.EphemeralTrue:
		return model.EphemeralPortState_ephemeralTrue
	case network.EphemeralFalse:
		return model.EphemeralPortState_ephemeralFalse
	default:
		return model.EphemeralPortState_ephemeralUnspecified
	}
}

func formatIPTranslation(ct *network.IPTranslation, ipc ipCache) *model.IPTranslation {
	if ct == nil {
		return nil
	}

	return &model.IPTranslation{
		ReplSrcIP:   ipc.Get(ct.ReplSrcIP),
		ReplDstIP:   ipc.Get(ct.ReplDstIP),
		ReplSrcPort: int32(ct.ReplSrcPort),
		ReplDstPort: int32(ct.ReplDstPort),
	}
}

func formatRouteIdx(v *network.Via, routes map[string]RouteIdx) int32 {
	if v == nil || routes == nil {
		return -1
	}

	if len(routes) == maxRoutes {
		return -1
	}

	k := routeKey(v)
	if len(k) == 0 {
		return -1
	}

	if idx, ok := routes[k]; ok {
		return idx.Idx
	}

	routes[k] = RouteIdx{
		Idx:   int32(len(routes)),
		Route: model.Route{Subnet: &model.Subnet{Alias: v.Subnet.Alias}},
	}

	return int32(len(routes)) - 1
}

func routeKey(v *network.Via) string {
	return v.Subnet.Alias
}

func formatTags(c network.ConnectionStats, tagsSet *network.TagsSet, connDynamicTags map[string]struct{}) (tagsIdx []uint32, checksum uint32) {
	mm := murmur3.New32()
	for _, tag := range network.GetStaticTags(c.StaticTags) {
		mm.Reset()
		_, _ = mm.Write(unsafeStringSlice(tag))
		checksum ^= mm.Sum32()
		tagsIdx = append(tagsIdx, tagsSet.Add(tag))
	}

	// Dynamic tags
	for tag := range connDynamicTags {
		mm.Reset()
		_, _ = mm.Write(unsafeStringSlice(tag))
		checksum ^= mm.Sum32()
		tagsIdx = append(tagsIdx, tagsSet.Add(tag))
	}

	// other tags, e.g., from process env vars like DD_ENV, etc.
	for tag := range c.Tags {
		mm.Reset()
		_, _ = mm.Write(unsafeStringSlice(tag))
		checksum ^= mm.Sum32()
		tagsIdx = append(tagsIdx, tagsSet.Add(tag))
	}

	return
}

func unsafeStringSlice(key string) []byte {
	if len(key) == 0 {
		return nil
	}
	// Reinterpret the string as bytes. This is safe because we don't write into the byte array.
	sh := (*reflect.StringHeader)(unsafe.Pointer(&key))
	return unsafe.Slice((*byte)(unsafe.Pointer(sh.Data)), len(key))
}
