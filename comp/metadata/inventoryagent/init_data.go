// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package inventoryagent

import (
	"context"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/flavor"
	"github.com/DataDog/datadog-agent/pkg/util/hostname"
	"github.com/DataDog/datadog-agent/pkg/util/installinfo"
	"github.com/DataDog/datadog-agent/pkg/util/scrubber"
	"github.com/DataDog/datadog-agent/pkg/version"
)

var (
	// for testing
	installinfoGet = installinfo.Get
)

func (ia *inventoryagent) initData() {
	clean := func(s string) string {
		// Errors come from internal use of a Reader interface. Since we are reading from a buffer, no errors
		// are possible.
		cleanBytes, _ := scrubber.ScrubBytes([]byte(s))
		return string(cleanBytes)
	}

	cfgSlice := func(name string) []string {
		if ia.conf.IsSet(name) {
			ss := ia.conf.GetStringSlice(name)
			rv := make([]string, len(ss))
			for i, s := range ss {
				rv[i] = clean(s)
			}
			return rv
		}
		return []string{}
	}

	tool := "undefined"
	toolVersion := ""
	installerVersion := ""

	install, err := installinfoGet(ia.conf)
	if err == nil {
		tool = install.Tool
		toolVersion = install.ToolVersion
		installerVersion = install.InstallerVersion
	}
	ia.data["install_method_tool"] = tool
	ia.data["install_method_tool_version"] = toolVersion
	ia.data["install_method_installer_version"] = installerVersion

	data, err := hostname.GetWithProvider(context.Background())
	if err != nil && data.Provider != "" && !data.FromFargate() {
		ia.data["hostname_source"] = data.Provider
	}

	ia.data["agent_version"] = version.AgentVersion
	ia.data["flavor"] = flavor.GetFlavor()

	ia.data["config_apm_dd_url"] = clean(ia.conf.GetString("apm_config.apm_dd_url"))
	ia.data["config_dd_url"] = clean(ia.conf.GetString("dd_url"))
	ia.data["config_site"] = clean(ia.conf.GetString("dd_site"))
	ia.data["config_logs_dd_url"] = clean(ia.conf.GetString("logs_config.logs_dd_url"))
	ia.data["config_logs_socks5_proxy_address"] = clean(ia.conf.GetString("logs_config.socks5_proxy_address"))
	ia.data["config_no_proxy"] = cfgSlice("proxy.no_proxy")
	ia.data["config_process_dd_url"] = clean(ia.conf.GetString("process_config.process_dd_url"))
	ia.data["config_proxy_http"] = clean(ia.conf.GetString("proxy.http"))
	ia.data["config_proxy_https"] = clean(ia.conf.GetString("proxy.https"))

	ia.data["feature_fips_enabled"] = ia.conf.GetBool("fips.enabled")
	ia.data["feature_logs_enabled"] = ia.conf.GetBool("logs_enabled")
	ia.data["feature_cspm_enabled"] = ia.conf.GetBool("compliance_config.enabled")
	ia.data["feature_apm_enabled"] = ia.conf.GetBool("apm_config.enabled")
	ia.data["feature_imdsv2_enabled"] = ia.conf.GetBool("ec2_prefer_imdsv2")
	ia.data["feature_dynamic_instrumentation_enabled"] = config.SystemProbe.GetBool("dynamic_instrumentation.enabled")
	ia.data["feature_remote_configuration_enabled"] = ia.conf.GetBool("remote_configuration.enabled")

	ia.data["feature_cws_enabled"] = config.SystemProbe.GetBool("runtime_security_config.enabled")
	ia.data["feature_cws_network_enabled"] = config.SystemProbe.GetBool("event_monitoring_config.network.enabled")
	ia.data["feature_cws_security_profiles_enabled"] = config.SystemProbe.GetBool("runtime_security_config.activity_dump.enabled")
	ia.data["feature_cws_remote_config_enabled"] = config.SystemProbe.GetBool("runtime_security_config.remote_configuration.enabled")

	ia.data["feature_process_enabled"] = ia.conf.GetBool("process_config.process_collection.enabled")
	ia.data["feature_process_language_detection_enabled"] = ia.conf.GetBool("language_detection.enabled")
	ia.data["feature_processes_container_enabled"] = ia.conf.GetBool("process_config.container_collection.enabled")

	ia.data["feature_networks_enabled"] = config.SystemProbe.GetBool("network_config.enabled")
	ia.data["feature_networks_http_enabled"] = config.SystemProbe.GetBool("service_monitoring_config.enable_http_monitoring")
	ia.data["feature_networks_https_enabled"] = config.SystemProbe.GetBool("service_monitoring_config.tls.native.enabled")

	ia.data["feature_usm_enabled"] = config.SystemProbe.GetBool("service_monitoring_config.enabled")
	ia.data["feature_usm_kafka_enabled"] = config.SystemProbe.GetBool("data_streams_config.enabled")
	ia.data["feature_usm_java_tls_enabled"] = config.SystemProbe.GetBool("service_monitoring_config.tls.java.enabled")
	ia.data["feature_usm_http2_enabled"] = config.SystemProbe.GetBool("service_monitoring_config.enable_http2_monitoring")
	ia.data["feature_usm_istio_enabled"] = config.SystemProbe.GetBool("service_monitoring_config.tls.istio.enabled")
	ia.data["feature_usm_http_by_status_code_enabled"] = config.SystemProbe.GetBool("service_monitoring_config.enable_http_stats_by_status_code")
	ia.data["feature_usm_go_tls_enabled"] = config.SystemProbe.GetBool("service_monitoring_config.tls.go.enabled")

	ia.data["feature_tcp_queue_length_enabled"] = config.SystemProbe.GetBool("system_probe_config.enable_tcp_queue_length")
	ia.data["feature_oom_kill_enabled"] = config.SystemProbe.GetBool("system_probe_config.enable_oom_kill")
	ia.data["feature_windows_crash_detection_enabled"] = config.SystemProbe.GetBool("windows_crash_detection.enabled")

	ia.data["system_probe_core_enabled"] = config.SystemProbe.GetBool("system_probe_config.enable_co_re")
	ia.data["system_probe_runtime_compilation_enabled"] = config.SystemProbe.GetBool("system_probe_config.enable_runtime_compiler")
	ia.data["system_probe_kernel_headers_download_enabled"] = config.SystemProbe.GetBool("system_probe_config.enable_kernel_header_download")
	ia.data["system_probe_prebuilt_fallback_enabled"] = config.SystemProbe.GetBool("system_probe_config.allow_precompiled_fallback")
	ia.data["system_probe_telemetry_enabled"] = config.SystemProbe.GetBool("system_probe_config.telemetry_enabled")
	ia.data["system_probe_max_connections_per_message"] = config.SystemProbe.GetInt("system_probe_config.max_conns_per_message")
	ia.data["system_probe_track_tcp_4_connections"] = config.SystemProbe.GetBool("network_config.collect_tcp_v4")
	ia.data["system_probe_track_tcp_6_connections"] = config.SystemProbe.GetBool("network_config.collect_tcp_v6")
	ia.data["system_probe_track_udp_4_connections"] = config.SystemProbe.GetBool("network_config.collect_udp_v4")
	ia.data["system_probe_track_udp_6_connections"] = config.SystemProbe.GetBool("network_config.collect_udp_v6")
	ia.data["system_probe_protocol_classification_enabled"] = config.SystemProbe.GetBool("network_config.enable_protocol_classification")
	ia.data["system_probe_gateway_lookup_enabled"] = config.SystemProbe.GetBool("network_config.enable_gateway_lookup")
	ia.data["system_probe_root_namespace_enabled"] = config.SystemProbe.GetBool("network_config.enable_root_netns")
}
