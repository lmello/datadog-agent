// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package model

// EventType describes the type of an event sent from the kernel
type EventType uint32

const (
	// UnknownEventType unknown event
	UnknownEventType EventType = iota
	// FileOpenEventType File open event
	FileOpenEventType
	// FileMkdirEventType Folder creation event
	FileMkdirEventType
	// FileLinkEventType Hard link creation event
	FileLinkEventType
	// FileRenameEventType File or folder rename event
	FileRenameEventType
	// FileUnlinkEventType Unlink event
	FileUnlinkEventType
	// FileRmdirEventType Rmdir event
	FileRmdirEventType
	// FileChmodEventType Chmod event
	FileChmodEventType
	// FileChownEventType Chown event
	FileChownEventType
	// FileUtimesEventType Utime event
	FileUtimesEventType
	// FileSetXAttrEventType Setxattr event
	FileSetXAttrEventType
	// FileRemoveXAttrEventType Removexattr event
	FileRemoveXAttrEventType
	// FileMountEventType Mount event
	FileMountEventType
	// FileUmountEventType Umount event
	FileUmountEventType
	// ForkEventType Fork event
	ForkEventType
	// ExecEventType Exec event
	ExecEventType
	// ExitEventType Exit event
	ExitEventType
	// InvalidateDentryEventType Dentry invalidated event (DEPRECATED)
	InvalidateDentryEventType
	// SetuidEventType setuid event
	SetuidEventType
	// SetgidEventType setgid event
	SetgidEventType
	// CapsetEventType capset event
	CapsetEventType
	// ArgsEnvsEventType args and envs event
	ArgsEnvsEventType
	// MountReleasedEventType sent when a mount point is released
	MountReleasedEventType
	// SELinuxEventType selinux event
	SELinuxEventType
	// BPFEventType bpf event
	BPFEventType
	// PTraceEventType PTrace event
	PTraceEventType
	// MMapEventType MMap event
	MMapEventType
	// MProtectEventType MProtect event
	MProtectEventType
	// LoadModuleEventType LoadModule event
	LoadModuleEventType
	// UnloadModuleEventType UnloadModule evnt
	UnloadModuleEventType
	// SignalEventType Signal event
	SignalEventType
	// SpliceEventType Splice event
	SpliceEventType
	// CgroupTracingEventType is sent when a new cgroup is being traced
	CgroupTracingEventType
	// DNSEventType DNS event
	DNSEventType
	// NetDeviceEventType is sent for events on net devices
	NetDeviceEventType
	// VethPairEventType is sent when a new veth pair is created
	VethPairEventType
	// BindEventType Bind event
	BindEventType
	// UnshareMountNsEventType is sent when a new mount is created from a mount namespace copy
	UnshareMountNsEventType
	// SyscallsEventType Syscalls event
	SyscallsEventType
	// AnomalyDetectionSyscallEventType Anomaly Detection Syscall event
	AnomalyDetectionSyscallEventType
	// MaxKernelEventType is used internally to get the maximum number of kernel events.
	MaxKernelEventType

	// FirstEventType is the first valid event type
	FirstEventType = FileOpenEventType

	// LastEventType is the last valid event type
	LastEventType = SyscallsEventType

	// FirstDiscarderEventType first event that accepts discarders
	FirstDiscarderEventType = FileOpenEventType

	// LastDiscarderEventType last event that accepts discarders
	LastDiscarderEventType = FileRemoveXAttrEventType

	// LastApproverEventType is the last event that accepts approvers
	LastApproverEventType = SpliceEventType

	// CustomLostReadEventType is the custom event used to report lost events detected in user space
	CustomLostReadEventType = iota
	// CustomLostWriteEventType is the custom event used to report lost events detected in kernel space
	CustomLostWriteEventType
	// CustomRulesetLoadedEventType is the custom event used to report that a new ruleset was loaded
	CustomRulesetLoadedEventType
	// CustomForkBombEventType is the custom event used to report the detection of a fork bomb
	CustomForkBombEventType
	// CustomTruncatedParentsEventType is the custom event used to report that the parents of a path were truncated
	CustomTruncatedParentsEventType
	// CustomSelfTestEventType is the custom event used to report the results of a self test run
	CustomSelfTestEventType
	// MaxAllEventType is used internally to get the maximum number of events.
	MaxAllEventType
)

func (t EventType) String() string {
	switch t {
	case FileOpenEventType:
		return "open"
	case FileMkdirEventType:
		return "mkdir"
	case FileLinkEventType:
		return "link"
	case FileRenameEventType:
		return "rename"
	case FileUnlinkEventType:
		return "unlink"
	case FileRmdirEventType:
		return "rmdir"
	case FileChmodEventType:
		return "chmod"
	case FileChownEventType:
		return "chown"
	case FileUtimesEventType:
		return "utimes"
	case FileMountEventType:
		return "mount"
	case FileUmountEventType:
		return "umount"
	case FileSetXAttrEventType:
		return "setxattr"
	case FileRemoveXAttrEventType:
		return "removexattr"
	case ForkEventType:
		return "fork"
	case ExecEventType:
		return "exec"
	case ExitEventType:
		return "exit"
	case InvalidateDentryEventType:
		return "invalidate_dentry"
	case SetuidEventType:
		return "setuid"
	case SetgidEventType:
		return "setgid"
	case CapsetEventType:
		return "capset"
	case ArgsEnvsEventType:
		return "args_envs"
	case MountReleasedEventType:
		return "mount_released"
	case SELinuxEventType:
		return "selinux"
	case BPFEventType:
		return "bpf"
	case PTraceEventType:
		return "ptrace"
	case MMapEventType:
		return "mmap"
	case MProtectEventType:
		return "mprotect"
	case LoadModuleEventType:
		return "load_module"
	case UnloadModuleEventType:
		return "unload_module"
	case SignalEventType:
		return "signal"
	case SpliceEventType:
		return "splice"
	case CgroupTracingEventType:
		return "cgroup_tracing"
	case DNSEventType:
		return "dns"
	case NetDeviceEventType:
		return "net_device"
	case VethPairEventType:
		return "veth_pair"
	case BindEventType:
		return "bind"
	case UnshareMountNsEventType:
		return "unshare_mntns"
	case SyscallsEventType:
		return "syscalls"
	case AnomalyDetectionSyscallEventType:
		return "anomaly_detection_syscall"

	case CustomLostReadEventType:
		return "lost_events_read"
	case CustomLostWriteEventType:
		return "lost_events_write"
	case CustomRulesetLoadedEventType:
		return "ruleset_loaded"
	case CustomForkBombEventType:
		return "fork_bomb"
	case CustomTruncatedParentsEventType:
		return "truncated_parents"
	case CustomSelfTestEventType:
		return "self_test"
	default:
		return "unknown"
	}
}
