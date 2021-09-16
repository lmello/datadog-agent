// Code generated by cmd/cgo -godefs; DO NOT EDIT.
// cgo -godefs -- -fsigned-char altkprobe_types.go

package ebpf

type Tuple struct {
	Family   uint8
	Protocol uint8
	Sport    uint16
	Dport    uint16
	Saddr    [16]uint8
	Daddr    [16]uint8
	Tgid     uint32
}
type FlowStats struct {
	Last_update uint64
	Sent_bytes  uint64
	Recv_bytes  uint64
}
type SocketInfo struct {
	Ns        uint64
	Tgid      uint32
	Netns     uint32
	Direction uint8
	Family    uint8
	Protocol  uint8
	Pad_cgo_0 [5]byte
}

type TCPSockStats struct {
	Retransmits       uint32
	Rtt               uint32
	Rtt_var           uint32
	State_transitions uint16
	Pad_cgo_0         [2]byte
}
type TCPFlow struct {
	Tup   Tuple
	Stats FlowStats
}
type TCPFlowKey struct {
	Skp       uint64
	Tgid      uint32
	Pad_cgo_0 [4]byte
}

type TCPCloseEvent struct {
	Skp      uint64
	Flow     TCPFlow
	Skinfo   SocketInfo
	Tcpstats TCPSockStats
}
type UDPCloseEvent struct {
	Skp    uint64
	Skinfo SocketInfo
}
