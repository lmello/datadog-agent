// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux_bpf || (windows && npm)

package tracer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	nethttp "net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cihub/seelog"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	ddconfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/ebpf/ebpftest"
	"github.com/DataDog/datadog-agent/pkg/network"
	"github.com/DataDog/datadog-agent/pkg/network/config"
	"github.com/DataDog/datadog-agent/pkg/process/util"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

var (
	clientMessageSize = 2 << 8
	serverMessageSize = 2 << 14
	payloadSizesTCP   = []int{2 << 5, 2 << 8, 2 << 10, 2 << 12, 2 << 14, 2 << 15}
	payloadSizesUDP   = []int{2 << 5, 2 << 8, 2 << 12, 2 << 14}
)

func TestMain(m *testing.M) {
	logLevel := os.Getenv("DD_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "warn"
	}
	log.SetupLogger(seelog.Default, logLevel)
	platformInit()
	os.Exit(m.Run())
}

type TracerSuite struct {
	suite.Suite
}

func TestTracerSuite(t *testing.T) {
	ebpftest.TestBuildModes(t, ebpftest.SupportedBuildModes(), "", func(t *testing.T) {
		suite.Run(t, new(TracerSuite))
	})
}

func isFentry() bool {
	fentryTests := os.Getenv("NETWORK_TRACER_FENTRY_TESTS")
	return fentryTests == "true"
}

func setupTracer(t testing.TB, cfg *config.Config) *Tracer {
	if isFentry() {
		ddconfig.SetFeatures(t, ddconfig.ECSFargate)
		// protocol classification not yet supported on fargate
		cfg.ProtocolClassificationEnabled = false
	}

	tr, err := NewTracer(cfg)
	require.NoError(t, err)
	t.Cleanup(tr.Stop)

	initTracerState(t, tr)
	return tr
}

func (s *TracerSuite) TestGetStats() {
	t := s.T()
	httpSupported := httpSupported()
	linuxExpected := map[string]interface{}{}
	err := json.Unmarshal([]byte(`{
      "state": {
        "closed_conn_dropped": 0,
		"conn_dropped": 0
      },
      "tracer": {
        "runtime": {
          "runtime_compilation_enabled": 0
        }
      }
    }`), &linuxExpected)
	require.NoError(t, err)

	rcExceptions := map[string]interface{}{}
	err = json.Unmarshal([]byte(`{
      "conntrack": {
        "evicts_total": 0,
        "orphan_size": 0
      }}`), &rcExceptions)
	require.NoError(t, err)

	expected := linuxExpected
	if runtime.GOOS == "windows" {
		expected = map[string]interface{}{
			"state": map[string]interface{}{},
		}
	}

	for _, enableEbpfConntracker := range []bool{true, false} {
		t.Run(fmt.Sprintf("ebpf conntracker %v", enableEbpfConntracker), func(t *testing.T) {
			cfg := testConfig()
			cfg.EnableHTTPMonitoring = true
			cfg.EnableEbpfConntracker = enableEbpfConntracker
			tr := setupTracer(t, cfg)

			<-time.After(time.Second)

			getConnections(t, tr)
			actual, _ := tr.getStats()

			for section, entries := range expected {
				if section == "usm" && !httpSupported {
					// HTTP stats not supported on some systems
					continue
				}
				require.Contains(t, actual, section, "missing section from telemetry map: %s", section)
				for name := range entries.(map[string]interface{}) {
					if cfg.EnableRuntimeCompiler || cfg.EnableEbpfConntracker {
						if sec, ok := rcExceptions[section]; ok {
							if _, ok := sec.(map[string]interface{})[name]; ok {
								continue
							}
						}
					}
					assert.Contains(t, actual[section], name, "%s actual is missing %s", section, name)
				}
			}
		})
	}
}

func (s *TracerSuite) TestTCPSendAndReceive() {
	t := s.T()
	tr := setupTracer(t, testConfig())

	// Create TCP Server which, for every line, sends back a message with size=serverMessageSize
	server := NewTCPServer(func(c net.Conn) {
		r := bufio.NewReader(c)
		for {
			_, err := r.ReadBytes(byte('\n'))
			c.Write(genPayload(serverMessageSize))
			if err != nil { // indicates that EOF has been reached,
				break
			}
		}
		c.Close()
	})
	t.Cleanup(server.Shutdown)
	err := server.Run()
	require.NoError(t, err)

	c, err := net.DialTimeout("tcp", server.address, 50*time.Millisecond)
	require.NoError(t, err)
	defer c.Close()

	// Connect to server 10 times
	wg := new(errgroup.Group)
	for i := 0; i < 10; i++ {
		wg.Go(func() error {
			// Write clientMessageSize to server, and read response
			if _, err = c.Write(genPayload(clientMessageSize)); err != nil {
				return err
			}

			r := bufio.NewReader(c)
			r.ReadBytes(byte('\n'))
			return nil
		})
	}

	err = wg.Wait()
	require.NoError(t, err)

	// Iterate through active connections until we find connection created above, and confirm send + recv counts
	connections := getConnections(t, tr)

	conn, ok := findConnection(c.LocalAddr(), c.RemoteAddr(), connections)
	require.True(t, ok)
	m := conn.Monotonic
	assert.Equal(t, 10*clientMessageSize, int(m.SentBytes))
	assert.Equal(t, 10*serverMessageSize, int(m.RecvBytes))
	assert.Equal(t, 0, int(m.Retransmits))
	assert.Equal(t, os.Getpid(), int(conn.Pid))
	assert.Equal(t, addrPort(server.address), int(conn.DPort))
	assert.Equal(t, network.OUTGOING, conn.Direction)
	assert.True(t, conn.IntraHost)
}

func (s *TracerSuite) TestTCPShortLived() {
	t := s.T()
	// Enable BPF-based system probe
	cfg := testConfig()
	cfg.TCPClosedTimeout = 10 * time.Millisecond
	tr := setupTracer(t, cfg)

	// Create TCP Server which sends back serverMessageSize bytes
	server := NewTCPServer(func(c net.Conn) {
		r := bufio.NewReader(c)
		r.ReadBytes(byte('\n'))
		c.Write(genPayload(serverMessageSize))
		c.Close()
	})
	t.Cleanup(server.Shutdown)
	require.NoError(t, server.Run())

	// Connect to server
	c, err := net.DialTimeout("tcp", server.address, 50*time.Millisecond)
	require.NoError(t, err)

	// Write clientMessageSize to server, and read response
	_, err = c.Write(genPayload(clientMessageSize))
	require.NoError(t, err)
	r := bufio.NewReader(c)
	r.ReadBytes(byte('\n'))

	// Explicitly close this TCP connection
	c.Close()

	var conn *network.ConnectionStats
	require.Eventually(t, func() bool {
		var ok bool
		conn, ok = findConnection(c.LocalAddr(), c.RemoteAddr(), getConnections(t, tr))
		return ok
	}, 3*time.Second, time.Second, "connection not found")

	m := conn.Monotonic
	assert.Equal(t, clientMessageSize, int(m.SentBytes))
	assert.Equal(t, serverMessageSize, int(m.RecvBytes))
	assert.Equal(t, 0, int(m.Retransmits))
	assert.Equal(t, os.Getpid(), int(conn.Pid))
	assert.Equal(t, addrPort(server.address), int(conn.DPort))
	assert.Equal(t, network.OUTGOING, conn.Direction)
	assert.True(t, conn.IntraHost)

	// Verify the short lived connection is accounting for both TCP_ESTABLISHED and TCP_CLOSED events
	assert.Equal(t, uint32(1), m.TCPEstablished)
	assert.Equal(t, uint32(1), m.TCPClosed)

	_, ok := findConnection(c.LocalAddr(), c.RemoteAddr(), getConnections(t, tr))
	assert.False(t, ok)
}

func (s *TracerSuite) TestTCPOverIPv6() {
	t := s.T()
	t.SkipNow()
	cfg := testConfig()
	cfg.CollectTCPv6Conns = true
	tr := setupTracer(t, cfg)

	ln, err := net.Listen("tcp6", ":0")
	require.NoError(t, err)

	doneChan := make(chan struct{})
	go func(done chan struct{}) {
		<-done
		ln.Close()
	}(doneChan)

	// Create TCP Server which sends back serverMessageSize bytes
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			r := bufio.NewReader(c)
			r.ReadBytes(byte('\n'))
			c.Write(genPayload(serverMessageSize))
			c.Close()
		}
	}()

	// Connect to server
	c, err := net.DialTimeout("tcp6", ln.Addr().String(), 50*time.Millisecond)
	require.NoError(t, err)

	// Write clientMessageSize to server, and read response
	if _, err = c.Write(genPayload(clientMessageSize)); err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(c)
	r.ReadBytes(byte('\n'))

	connections := getConnections(t, tr)

	conn, ok := findConnection(c.LocalAddr(), c.RemoteAddr(), connections)
	require.True(t, ok)
	m := conn.Monotonic
	assert.Equal(t, clientMessageSize, int(m.SentBytes))
	assert.Equal(t, serverMessageSize, int(m.RecvBytes))
	assert.Equal(t, 0, int(m.Retransmits))
	assert.Equal(t, os.Getpid(), int(conn.Pid))
	assert.Equal(t, ln.Addr().(*net.TCPAddr).Port, int(conn.DPort))
	assert.Equal(t, network.OUTGOING, conn.Direction)
	assert.True(t, conn.IntraHost)

	doneChan <- struct{}{}
}

func (s *TracerSuite) TestTCPCollectionDisabled() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("Test disabled on Windows")
	}
	// Enable BPF-based system probe with TCP disabled
	cfg := testConfig()
	cfg.CollectTCPv4Conns = false
	cfg.CollectTCPv6Conns = false
	tr := setupTracer(t, cfg)

	// Create TCP Server which sends back serverMessageSize bytes
	server := NewTCPServer(func(c net.Conn) {
		r := bufio.NewReader(c)
		r.ReadBytes(byte('\n'))
		c.Write(genPayload(serverMessageSize))
		c.Close()
	})

	t.Cleanup(server.Shutdown)
	require.NoError(t, server.Run())

	// Connect to server
	c, err := net.DialTimeout("tcp", server.address, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Write clientMessageSize to server, and read response
	if _, err = c.Write(genPayload(clientMessageSize)); err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(c)
	r.ReadBytes(byte('\n'))

	connections := getConnections(t, tr)

	// Confirm that we could not find connection created above
	_, ok := findConnection(c.LocalAddr(), c.RemoteAddr(), connections)
	require.False(t, ok)
}

func (s *TracerSuite) TestTCPConnsReported() {
	t := s.T()
	// Setup
	cfg := testConfig()
	cfg.CollectTCPv4Conns = true
	cfg.CollectTCPv6Conns = true
	tr := setupTracer(t, cfg)

	processedChan := make(chan struct{})
	server := NewTCPServer(func(c net.Conn) {
		c.Close()
		close(processedChan)
	})
	t.Cleanup(server.Shutdown)
	require.NoError(t, server.Run())

	// Connect to server
	c, err := net.DialTimeout("tcp", server.address, 50*time.Millisecond)
	require.NoError(t, err)
	defer c.Close()
	<-processedChan

	// Test
	connections := getConnections(t, tr)
	// Server-side
	_, ok := findConnection(c.RemoteAddr(), c.LocalAddr(), connections)
	require.True(t, ok)
	// Client-side
	_, ok = findConnection(c.LocalAddr(), c.RemoteAddr(), connections)
	require.True(t, ok)
}

func (s *TracerSuite) TestUDPSendAndReceive() {
	t := s.T()
	cfg := testConfig()
	tr := setupTracer(t, cfg)

	t.Run("v4", func(t *testing.T) {
		if !testConfig().CollectUDPv4Conns {
			t.Skip("UDPv4 disabled")
		}
		t.Run("fixed port", func(t *testing.T) {
			testUDPSendAndReceive(t, tr, "127.0.0.1:8081")
		})
	})
	t.Run("v6", func(t *testing.T) {
		if !testConfig().CollectUDPv6Conns {
			t.Skip("UDPv6 disabled")
		}
		t.Run("fixed port", func(t *testing.T) {
			testUDPSendAndReceive(t, tr, "[::1]:8081")
		})
	})
}

func testUDPSendAndReceive(t *testing.T, tr *Tracer, addr string) {
	tr.removeClient(clientID)

	server := &UDPServer{
		address: addr,
		onMessage: func(buf []byte, n int) []byte {
			return genPayload(serverMessageSize)
		},
	}

	err := server.Run(clientMessageSize)
	require.NoError(t, err)
	t.Cleanup(server.Shutdown)

	initTracerState(t, tr)

	// Connect to server
	c, err := net.DialTimeout("udp", server.address, 50*time.Millisecond)
	require.NoError(t, err)
	defer c.Close()

	// Write clientMessageSize to server, and read response
	_, err = c.Write(genPayload(clientMessageSize))
	require.NoError(t, err)

	_, err = c.Read(make([]byte, serverMessageSize))
	require.NoError(t, err)

	// Iterate through active connections until we find connection created above, and confirm send + recv counts
	connections := getConnections(t, tr)

	incoming, ok := findConnection(c.RemoteAddr(), c.LocalAddr(), connections)
	if assert.True(t, ok, "unable to find incoming connection") {
		assert.Equal(t, network.INCOMING, incoming.Direction)

		// make sure the inverse values are seen for the other message
		assert.Equal(t, serverMessageSize, int(incoming.Monotonic.SentBytes), "incoming sent")
		assert.Equal(t, clientMessageSize, int(incoming.Monotonic.RecvBytes), "incoming recv")
		assert.True(t, incoming.IntraHost, "incoming intrahost")
	}

	outgoing, ok := findConnection(c.LocalAddr(), c.RemoteAddr(), connections)
	if assert.True(t, ok, "unable to find outgoing connection") {
		assert.Equal(t, network.OUTGOING, outgoing.Direction)

		assert.Equal(t, clientMessageSize, int(outgoing.Monotonic.SentBytes), "outgoing sent")
		assert.Equal(t, serverMessageSize, int(outgoing.Monotonic.RecvBytes), "outgoing recv")
		assert.True(t, outgoing.IntraHost, "outgoing intrahost")
	}
}

func (s *TracerSuite) TestUDPDisabled() {
	t := s.T()
	// Enable BPF-based system probe with UDP disabled
	cfg := testConfig()
	cfg.CollectUDPv4Conns = false
	cfg.CollectUDPv6Conns = false
	tr := setupTracer(t, cfg)

	// Create UDP Server which sends back serverMessageSize bytes
	server := &UDPServer{
		onMessage: func(b []byte, n int) []byte {
			return genPayload(serverMessageSize)
		},
	}

	err := server.Run(clientMessageSize)
	require.NoError(t, err)
	t.Cleanup(server.Shutdown)

	// Connect to server
	c, err := net.DialTimeout("udp", server.address, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// Write clientMessageSize to server, and read response
	if _, err = c.Write(genPayload(clientMessageSize)); err != nil {
		t.Fatal(err)
	}

	c.Read(make([]byte, serverMessageSize))

	// Iterate through active connections until we find connection created above, and confirm send + recv counts
	connections := getConnections(t, tr)

	_, ok := findConnection(c.LocalAddr(), c.RemoteAddr(), connections)
	require.False(t, ok)
}

func (s *TracerSuite) TestLocalDNSCollectionDisabled() {
	t := s.T()
	// Enable BPF-based system probe with DNS disabled (by default)
	config := testConfig()

	tr := setupTracer(t, config)

	// Connect to local DNS
	addr, err := net.ResolveUDPAddr("udp", "localhost:53")
	assert.NoError(t, err)

	cn, err := net.DialUDP("udp", nil, addr)
	assert.NoError(t, err)
	defer cn.Close()

	// Write anything
	_, err = cn.Write([]byte("test"))
	assert.NoError(t, err)

	// Iterate through active connections making sure there are no local DNS calls
	for _, c := range getConnections(t, tr).Conns {
		assert.False(t, isLocalDNS(c))
	}
}

func (s *TracerSuite) TestLocalDNSCollectionEnabled() {
	t := s.T()
	// Enable BPF-based system probe with DNS enabled
	cfg := testConfig()
	cfg.CollectLocalDNS = true
	cfg.CollectUDPv4Conns = true
	tr := setupTracer(t, cfg)

	// Connect to local DNS
	addr, err := net.ResolveUDPAddr("udp", "localhost:53")
	assert.NoError(t, err)

	cn, err := net.DialUDP("udp", nil, addr)
	assert.NoError(t, err)
	defer cn.Close()

	// Write anything
	_, err = cn.Write([]byte("test"))
	assert.NoError(t, err)

	found := false

	// Iterate through active connections making sure theres at least one connection
	for _, c := range getConnections(t, tr).Conns {
		found = found || isLocalDNS(c)
	}

	assert.True(t, found)
}

func isLocalDNS(c network.ConnectionStats) bool {
	return c.Source.String() == "127.0.0.1" && c.Dest.String() == "127.0.0.1" && c.DPort == 53
}

func (s *TracerSuite) TestShouldSkipExcludedConnection() {
	t := s.T()
	// exclude connections from 127.0.0.1:80
	cfg := testConfig()
	// exclude source SSH connections to make this pass in VM
	cfg.ExcludedSourceConnections = map[string][]string{"127.0.0.1": {"80"}, "*": {"22"}}
	cfg.ExcludedDestinationConnections = map[string][]string{"127.0.0.1": {"tcp 80"}}
	tr := setupTracer(t, cfg)

	// Connect to 127.0.0.1:80
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:80")
	assert.NoError(t, err)

	cn, err := net.DialUDP("udp", nil, addr)
	assert.NoError(t, err)
	defer cn.Close()

	// Write anything
	_, err = cn.Write([]byte("test"))
	assert.NoError(t, err)

	// Make sure we're not picking up 127.0.0.1:80
	cxs := getConnections(t, tr)
	for _, c := range cxs.Conns {
		assert.False(t, c.Source.String() == "127.0.0.1" && c.SPort == 80, "connection %s should be excluded", c)
		assert.False(t, c.Dest.String() == "127.0.0.1" && c.DPort == 80 && c.Type == network.TCP, "connection %s should be excluded", c)
	}

	// ensure one of the connections is UDP to 127.0.0.1:80
	assert.Condition(t, func() bool {
		for _, c := range cxs.Conns {
			if c.Dest.String() == "127.0.0.1" && c.DPort == 80 && c.Type == network.UDP {
				return true
			}
		}
		return false
	}, "Unable to find UDP connection to 127.0.0.1:80")
}

func (s *TracerSuite) TestShouldExcludeEmptyStatsConnection() {
	t := s.T()
	cfg := testConfig()
	tr := setupTracer(t, cfg)

	// Connect to 127.0.0.1:80
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:80")
	assert.NoError(t, err)

	cn, err := net.DialUDP("udp", nil, addr)
	assert.NoError(t, err)
	defer cn.Close()

	// Write anything
	_, err = cn.Write([]byte("test"))
	assert.NoError(t, err)

	var zeroConn network.ConnectionStats
	require.Eventually(t, func() bool {
		cxs := getConnections(t, tr)
		for _, c := range cxs.Conns {
			if c.Dest.String() == "127.0.0.1" && c.DPort == 80 {
				zeroConn = c
				return true
			}
		}
		return false
	}, 2*time.Second, time.Second)

	// next call should not have the same connection
	cxs := getConnections(t, tr)
	found := false
	for _, c := range cxs.Conns {
		if c.Source == zeroConn.Source && c.SPort == zeroConn.SPort &&
			c.Dest == zeroConn.Dest && c.DPort == zeroConn.DPort &&
			c.Pid == zeroConn.Pid {
			found = true
			break
		}
	}
	require.False(t, found, "empty connections should be filtered out")
}

func TestSkipConnectionDNS(t *testing.T) {
	t.Run("CollectLocalDNS disabled", func(t *testing.T) {
		tr := &Tracer{config: &config.Config{CollectLocalDNS: false}}
		assert.True(t, tr.shouldSkipConnection(&network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.1"),
			Dest:   util.AddressFromString("127.0.0.1"),
			SPort:  1000, DPort: 53,
		}))

		assert.False(t, tr.shouldSkipConnection(&network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.1"),
			Dest:   util.AddressFromString("127.0.0.1"),
			SPort:  1000, DPort: 8080,
		}))

		assert.True(t, tr.shouldSkipConnection(&network.ConnectionStats{
			Source: util.AddressFromString("::3f::45"),
			Dest:   util.AddressFromString("::1"),
			SPort:  53, DPort: 1000,
		}))

		assert.True(t, tr.shouldSkipConnection(&network.ConnectionStats{
			Source: util.AddressFromString("::3f::45"),
			Dest:   util.AddressFromString("::1"),
			SPort:  53, DPort: 1000,
		}))
	})

	t.Run("CollectLocalDNS disabled", func(t *testing.T) {
		tr := &Tracer{config: &config.Config{CollectLocalDNS: true}}

		assert.False(t, tr.shouldSkipConnection(&network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.1"),
			Dest:   util.AddressFromString("127.0.0.1"),
			SPort:  1000, DPort: 53,
		}))

		assert.False(t, tr.shouldSkipConnection(&network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.1"),
			Dest:   util.AddressFromString("127.0.0.1"),
			SPort:  1000, DPort: 8080,
		}))

		assert.False(t, tr.shouldSkipConnection(&network.ConnectionStats{
			Source: util.AddressFromString("::3f::45"),
			Dest:   util.AddressFromString("::1"),
			SPort:  53, DPort: 1000,
		}))

		assert.False(t, tr.shouldSkipConnection(&network.ConnectionStats{
			Source: util.AddressFromString("::3f::45"),
			Dest:   util.AddressFromString("::1"),
			SPort:  53, DPort: 1000,
		}))
	})
}

func findConnection(l, r net.Addr, c *network.Connections) (*network.ConnectionStats, bool) {
	res := network.FirstConnection(c, network.ByTuple(l, r))
	return res, res != nil
}

func searchConnections(c *network.Connections, predicate func(network.ConnectionStats) bool) []network.ConnectionStats {
	return network.FilterConnections(c, predicate)
}

func runBenchtests(b *testing.B, payloads []int, prefix string, f func(p int) func(*testing.B)) {
	for _, p := range payloads {
		name := strings.TrimSpace(strings.Join([]string{prefix, strconv.Itoa(p), "bytes"}, " "))
		b.Run(name, f(p))
	}
}

func BenchmarkUDPEcho(b *testing.B) {
	runBenchtests(b, payloadSizesUDP, "", benchEchoUDP)

	// Enable BPF-based system probe
	_ = setupTracer(b, testConfig())

	runBenchtests(b, payloadSizesUDP, "eBPF", benchEchoUDP)
}

func benchEchoUDP(size int) func(b *testing.B) {
	payload := genPayload(size)
	echoOnMessage := func(b []byte, n int) []byte {
		resp := make([]byte, len(b))
		copy(resp, b)
		return resp
	}

	return func(b *testing.B) {
		server := &UDPServer{onMessage: echoOnMessage}
		err := server.Run(size)
		require.NoError(b, err)
		defer server.Shutdown()

		c, err := net.DialTimeout("udp", server.address, 50*time.Millisecond)
		if err != nil {
			b.Fatal(err)
		}
		defer c.Close()
		r := bufio.NewReader(c)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Write(payload)
			buf := make([]byte, size)
			n, err := r.Read(buf)

			if err != nil || n != len(payload) || !bytes.Equal(payload, buf) {
				b.Fatalf("Sizes: %d, %d. Equal: %v. Error: %s", len(buf), len(payload), bytes.Equal(payload, buf), err)
			}
		}
		b.StopTimer()
	}
}

func BenchmarkTCPEcho(b *testing.B) {
	runBenchtests(b, payloadSizesTCP, "", benchEchoTCP)

	// Enable BPF-based system probe
	_ = setupTracer(b, testConfig())
	runBenchtests(b, payloadSizesTCP, "eBPF", benchEchoTCP)
}

func BenchmarkTCPSend(b *testing.B) {
	runBenchtests(b, payloadSizesTCP, "", benchSendTCP)

	// Enable BPF-based system probe
	_ = setupTracer(b, testConfig())
	runBenchtests(b, payloadSizesTCP, "eBPF", benchSendTCP)
}

func benchEchoTCP(size int) func(b *testing.B) {
	payload := genPayload(size)
	echoOnMessage := func(c net.Conn) {
		r := bufio.NewReader(c)
		for {
			buf, err := r.ReadBytes(byte('\n'))
			if err == io.EOF {
				c.Close()
				return
			}
			c.Write(buf)
		}
	}

	return func(b *testing.B) {
		server := NewTCPServer(echoOnMessage)
		b.Cleanup(server.Shutdown)
		require.NoError(b, server.Run())

		c, err := net.DialTimeout("tcp", server.address, 50*time.Millisecond)
		if err != nil {
			b.Fatal(err)
		}
		defer c.Close()
		r := bufio.NewReader(c)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Write(payload)
			buf, err := r.ReadBytes(byte('\n'))

			if err != nil || len(buf) != len(payload) || !bytes.Equal(payload, buf) {
				b.Fatalf("Sizes: %d, %d. Equal: %v. Error: %s", len(buf), len(payload), bytes.Equal(payload, buf), err)
			}
		}
		b.StopTimer()
	}
}

func benchSendTCP(size int) func(b *testing.B) {
	payload := genPayload(size)
	dropOnMessage := func(c net.Conn) {
		r := bufio.NewReader(c)
		for { // Drop all payloads received
			_, err := r.Discard(r.Buffered() + 1)
			if err == io.EOF {
				c.Close()
				return
			}
		}
	}

	return func(b *testing.B) {
		server := NewTCPServer(dropOnMessage)
		b.Cleanup(server.Shutdown)
		require.NoError(b, server.Run())

		c, err := net.DialTimeout("tcp", server.address, 50*time.Millisecond)
		if err != nil {
			b.Fatal(err)
		}
		defer c.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ { // Send-heavy workload
			_, err := c.Write(payload)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
	}
}

type TCPServer struct {
	address   string
	network   string
	onMessage func(c net.Conn)
	ln        net.Listener
}

func NewTCPServer(onMessage func(c net.Conn)) *TCPServer {
	return NewTCPServerOnAddress("127.0.0.1:0", onMessage)
}

func NewTCPServerOnAddress(addr string, onMessage func(c net.Conn)) *TCPServer {
	return &TCPServer{
		address:   addr,
		onMessage: onMessage,
	}
}

func (t *TCPServer) Run() error {
	networkType := "tcp"
	if t.network != "" {
		networkType = t.network
	}
	ln, err := net.Listen(networkType, t.address)
	if err != nil {
		return err
	}
	t.ln = ln
	t.address = ln.Addr().String()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go t.onMessage(conn)
		}
	}()

	return nil
}

func (t *TCPServer) Shutdown() {
	if t.ln != nil {
		_ = t.ln.Close()
		t.ln = nil
	}
}

type UDPServer struct {
	network   string
	address   string
	lc        *net.ListenConfig
	onMessage func(b []byte, n int) []byte
	ln        net.PacketConn
}

func (s *UDPServer) Run(payloadSize int) error {
	networkType := "udp"
	if s.network != "" {
		networkType = s.network
	}
	var err error
	var ln net.PacketConn
	if s.lc != nil {
		ln, err = s.lc.ListenPacket(context.Background(), networkType, s.address)
	} else {
		ln, err = net.ListenPacket(networkType, s.address)
	}
	if err != nil {
		return err
	}

	s.ln = ln
	s.address = s.ln.LocalAddr().String()

	go func() {
		buf := make([]byte, payloadSize)
		for {
			n, addr, err := ln.ReadFrom(buf)
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					fmt.Printf("readfrom: %s\n", err)
				}
				return
			}
			ret := s.onMessage(buf, n)
			if ret != nil {
				_, err = s.ln.WriteTo(ret, addr)
				if err != nil {
					if !errors.Is(err, net.ErrClosed) {
						fmt.Printf("writeto: %s\n", err)
					}
					return
				}
			}
		}
	}()

	return nil
}

func (s *UDPServer) Shutdown() {
	if s.ln != nil {
		_ = s.ln.Close()
		s.ln = nil
	}
}

var letterBytes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func genPayload(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		if i == n-1 {
			b[i] = '\n'
		} else {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
	}
	return b
}

func addrPort(addr string) int {
	p, _ := strconv.Atoi(strings.Split(addr, ":")[1])
	return p
}

const clientID = "1"

func initTracerState(t testing.TB, tr *Tracer) {
	err := tr.RegisterClient(clientID)
	require.NoError(t, err)
}

func getConnections(t require.TestingT, tr *Tracer) *network.Connections {
	// Iterate through active connections until we find connection created above, and confirm send + recv counts
	connections, err := tr.GetActiveConnections(clientID)
	require.NoError(t, err)
	return connections
}

const (
	validDNSServer = "8.8.8.8"
)

func testDNSStats(t *testing.T, tr *Tracer, domain string, success, failure, timeout int, serverIP string) {
	tr.removeClient(clientID)
	initTracerState(t, tr)

	dnsServerAddr := &net.UDPAddr{IP: net.ParseIP(serverIP), Port: 53}

	queryMsg := new(dns.Msg)
	queryMsg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	queryMsg.RecursionDesired = true

	dnsClient := new(dns.Client)
	dnsConn, err := dnsClient.Dial(dnsServerAddr.String())
	require.NoError(t, err)
	defer dnsConn.Close()
	dnsClientAddr := dnsConn.LocalAddr().(*net.UDPAddr)
	_, _, err = dnsClient.ExchangeWithConn(queryMsg, dnsConn)

	if err != nil && timeout == 0 {
		t.Fatalf("Failed to get dns response %s\n", err.Error())
	}

	// Allow the DNS reply to be processed in the snooper
	time.Sleep(time.Millisecond * 500)

	// Iterate through active connections until we find connection created above, and confirm send + recv counts
	connections := getConnections(t, tr)
	conn, ok := findConnection(dnsClientAddr, dnsServerAddr, connections)
	require.True(t, ok)

	assert.Equal(t, queryMsg.Len(), int(conn.Monotonic.SentBytes))
	assert.Equal(t, os.Getpid(), int(conn.Pid))
	assert.Equal(t, dnsServerAddr.Port, int(conn.DPort))

	dnsKey, ok := network.DNSKey(conn)
	require.True(t, ok)

	dnsStats, ok := connections.DNSStats[dnsKey]
	require.True(t, ok)

	var total uint32
	var successfulResponses uint32
	var timeouts uint32
	for _, byDomain := range dnsStats {
		for _, byQueryType := range byDomain {
			successfulResponses += byQueryType.CountByRcode[uint32(0)]
			timeouts += byQueryType.Timeouts
			for _, count := range byQueryType.CountByRcode {
				total += count
			}
		}
	}

	failedResponses := total - successfulResponses

	// DNS Stats
	assert.Equal(t, uint32(success), successfulResponses)
	assert.Equal(t, uint32(failure), failedResponses)
	assert.Equal(t, uint32(timeout), timeouts)
}

func (s *TracerSuite) TestDNSStats() {
	t := s.T()
	cfg := testConfig()
	cfg.CollectDNSStats = true
	cfg.DNSTimeout = 1 * time.Second
	tr := setupTracer(t, cfg)
	t.Run("valid domain", func(t *testing.T) {
		testDNSStats(t, tr, "golang.org", 1, 0, 0, validDNSServer)
	})
	t.Run("invalid domain", func(t *testing.T) {
		testDNSStats(t, tr, "abcdedfg", 0, 1, 0, validDNSServer)
	})
	t.Run("timeout", func(t *testing.T) {
		testDNSStats(t, tr, "golang.org", 0, 0, 1, "1.2.3.4")
	})
}

func (s *TracerSuite) TestTCPEstablished() {
	t := s.T()
	// Ensure closed connections are flushed as soon as possible
	cfg := testConfig()
	cfg.TCPClosedTimeout = 500 * time.Millisecond

	tr := setupTracer(t, cfg)

	server := NewTCPServer(func(c net.Conn) {
		io.Copy(io.Discard, c)
		c.Close()
	})
	t.Cleanup(server.Shutdown)
	require.NoError(t, server.Run())

	c, err := net.DialTimeout("tcp", server.address, 50*time.Millisecond)
	require.NoError(t, err)

	laddr, raddr := c.LocalAddr(), c.RemoteAddr()
	c.Write([]byte("hello"))

	connections := getConnections(t, tr)
	conn, ok := findConnection(laddr, raddr, connections)

	require.True(t, ok)
	assert.Equal(t, uint32(1), conn.Last.TCPEstablished)
	assert.Equal(t, uint32(0), conn.Last.TCPClosed)

	c.Close()
	// Wait for the connection to be sent from the perf buffer
	time.Sleep(cfg.TCPClosedTimeout)

	connections = getConnections(t, tr)
	conn, ok = findConnection(laddr, raddr, connections)
	require.True(t, ok)
	assert.Equal(t, uint32(0), conn.Last.TCPEstablished)
	assert.Equal(t, uint32(1), conn.Last.TCPClosed)
}

func (s *TracerSuite) TestTCPEstablishedPreExistingConn() {
	t := s.T()
	server := NewTCPServer(func(c net.Conn) {
		io.Copy(io.Discard, c)
		c.Close()
	})
	t.Cleanup(server.Shutdown)
	require.NoError(t, server.Run())

	c, err := net.DialTimeout("tcp", server.address, 50*time.Millisecond)
	require.NoError(t, err)
	laddr, raddr := c.LocalAddr(), c.RemoteAddr()

	// Ensure closed connections are flushed as soon as possible
	cfg := testConfig()
	cfg.TCPClosedTimeout = 500 * time.Millisecond

	tr := setupTracer(t, cfg)

	c.Write([]byte("hello"))
	c.Close()
	// Wait for the connection to be sent from the perf buffer
	time.Sleep(cfg.TCPClosedTimeout)
	connections := getConnections(t, tr)
	conn, ok := findConnection(laddr, raddr, connections)

	require.True(t, ok)
	m := conn.Monotonic
	assert.Equal(t, uint32(0), m.TCPEstablished)
	assert.Equal(t, uint32(1), m.TCPClosed)
}

func (s *TracerSuite) TestUnconnectedUDPSendIPv4() {
	t := s.T()
	cfg := testConfig()
	tr := setupTracer(t, cfg)

	remotePort := rand.Int()%5000 + 15000
	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: remotePort}
	// Use ListenUDP instead of DialUDP to create a "connectionless" UDP connection
	conn, err := net.ListenUDP("udp4", nil)
	require.NoError(t, err)
	defer conn.Close()
	message := []byte("payload")
	bytesSent, err := conn.WriteTo(message, remoteAddr)
	require.NoError(t, err)

	connections := getConnections(t, tr)
	outgoing := searchConnections(connections, func(cs network.ConnectionStats) bool {
		return cs.DPort == uint16(remotePort)
	})

	require.Len(t, outgoing, 1)
	assert.Equal(t, bytesSent, int(outgoing[0].Monotonic.SentBytes))
}

func (s *TracerSuite) TestConnectedUDPSendIPv6() {
	t := s.T()
	cfg := testConfig()
	if !testConfig().CollectUDPv6Conns {
		t.Skip("UDPv6 disabled")
	}
	tr := setupTracer(t, cfg)

	remotePort := rand.Int()%5000 + 15000
	remoteAddr := &net.UDPAddr{IP: net.IPv6loopback, Port: remotePort}
	conn, err := net.DialUDP("udp6", nil, remoteAddr)
	require.NoError(t, err)
	defer conn.Close()
	message := []byte("payload")
	bytesSent, err := conn.Write(message)
	require.NoError(t, err)

	connections := getConnections(t, tr)
	outgoing := searchConnections(connections, func(cs network.ConnectionStats) bool {
		return cs.DPort == uint16(remotePort)
	})

	require.Len(t, outgoing, 1)
	assert.Equal(t, remoteAddr.IP.String(), outgoing[0].Dest.String())
	assert.Equal(t, bytesSent, int(outgoing[0].Monotonic.SentBytes))
}

func (s *TracerSuite) TestTCPDirection() {
	t := s.T()
	cfg := testConfig()
	tr := setupTracer(t, cfg)

	// Start an HTTP server on localhost:8080
	serverAddr := "127.0.0.1:8080"
	srv := &nethttp.Server{
		Addr: serverAddr,
		Handler: nethttp.HandlerFunc(func(w nethttp.ResponseWriter, req *nethttp.Request) {
			t.Logf("received http request from %s", req.RemoteAddr)
			io.Copy(io.Discard, req.Body)
			w.WriteHeader(200)
		}),
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}
	srv.SetKeepAlivesEnabled(false)
	go func() {
		_ = srv.ListenAndServe()
	}()
	defer srv.Shutdown(context.Background())

	// Allow the HTTP server time to get set up
	time.Sleep(time.Millisecond * 500)

	// Send a HTTP request to the test server
	client := new(nethttp.Client)
	resp, err := client.Get("http://" + serverAddr + "/test")
	require.NoError(t, err)
	resp.Body.Close()

	// Iterate through active connections until we find connection created above
	var outgoingConns []network.ConnectionStats
	var incomingConns []network.ConnectionStats
	require.Eventuallyf(t, func() bool {
		conns := getConnections(t, tr)
		if len(outgoingConns) == 0 {
			outgoingConns = searchConnections(conns, func(cs network.ConnectionStats) bool {
				return fmt.Sprintf("%s:%d", cs.Dest, cs.DPort) == serverAddr
			})
		}
		if len(incomingConns) == 0 {
			incomingConns = searchConnections(conns, func(cs network.ConnectionStats) bool {
				return fmt.Sprintf("%s:%d", cs.Source, cs.SPort) == serverAddr
			})
		}

		return len(outgoingConns) == 1 && len(incomingConns) == 1
	}, 3*time.Second, 10*time.Millisecond, "couldn't find incoming and outgoing http connections matching: %s", serverAddr)

	// Verify connection directions
	conn := outgoingConns[0]
	assert.Equal(t, conn.Direction, network.OUTGOING, "connection direction must be outgoing: %s", conn)
	conn = incomingConns[0]
	assert.Equal(t, conn.Direction, network.INCOMING, "connection direction must be incoming: %s", conn)
}
