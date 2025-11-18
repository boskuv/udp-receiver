package services

import (
	"net"
	"testing"
	"time"
)

func TestListenOnPorts(t *testing.T) {
	tests := []struct {
		name          string
		udpServices   map[string]int
		timeoutSec    int
		wantErr       bool
		validatePorts bool
	}{
		{
			name:          "single service",
			udpServices:   map[string]int{"test": 0}, // Use port 0 for automatic port assignment
			timeoutSec:    5,
			wantErr:       false,
			validatePorts: false,
		},
		{
			name:          "multiple services",
			udpServices:   map[string]int{"service1": 0, "service2": 0},
			timeoutSec:    5,
			wantErr:       false,
			validatePorts: false,
		},
		{
			name:        "no services",
			udpServices: map[string]int{},
			timeoutSec:  5,
			wantErr:     true,
		},
		{
			name:        "invalid port - negative",
			udpServices: map[string]int{"test": -1},
			timeoutSec:  5,
			wantErr:     true,
		},
		{
			name:        "invalid port - too large",
			udpServices: map[string]int{"test": 65536},
			timeoutSec:  5,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For tests that need actual ports, we'll use port 0 to get an available port
			services := make(map[string]int)
			for k, v := range tt.udpServices {
				if v == 0 && !tt.wantErr {
					// Find an available port
					addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
					if err != nil {
						t.Fatalf("Failed to resolve UDP addr: %v", err)
					}
					conn, err := net.ListenUDP("udp", addr)
					if err != nil {
						t.Fatalf("Failed to listen: %v", err)
					}
					port := conn.LocalAddr().(*net.UDPAddr).Port
					conn.Close()
					services[k] = port
				} else {
					services[k] = v
				}
			}

			conns, err := ListenOnPorts(services, tt.timeoutSec)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListenOnPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(conns) != len(services) {
					t.Errorf("ListenOnPorts() returned %d connections, want %d", len(conns), len(services))
				}

				// Clean up connections
				for _, conn := range conns {
					if conn.PacketConn != nil {
						conn.PacketConn.Close()
					}
				}
			}
		})
	}
}

func TestHandlePacket(t *testing.T) {
	// Create a test UDP listener
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP addr: %v", err)
	}

	serverConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer serverConn.Close()

	serverAddr := serverConn.LocalAddr().(*net.UDPAddr)

	// Create service connection
	serviceConns := []ServiceNetConnection{
		{
			ServiceName: "test-service",
			PacketConn:  serverConn,
		},
	}

	statusChan := make(chan ServiceNetStatus, 10)
	sleepTimeSec := 1
	answerTimeoutSec := 2

	// Start packet handler
	HandlePacket(sleepTimeSec, answerTimeoutSec, serviceConns, statusChan)

	// Send a test packet
	clientConn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer clientConn.Close()

	testData := []byte("test packet")
	_, err = clientConn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Wait for status update
	select {
	case status := <-statusChan:
		if status.ServiceName != "test-service" {
			t.Errorf("ServiceName = %v, want test-service", status.ServiceName)
		}
		if status.Status != 1 {
			t.Errorf("Status = %v, want 1", status.Status)
		}
	case <-time.After(3 * time.Second):
		t.Error("Timeout waiting for status update")
	}
}

func TestHandlePacket_Timeout(t *testing.T) {
	// Create a test UDP listener that won't receive packets
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP addr: %v", err)
	}

	serverConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer serverConn.Close()

	// Create service connection
	serviceConns := []ServiceNetConnection{
		{
			ServiceName: "test-service",
			PacketConn:  serverConn,
		},
	}

	statusChan := make(chan ServiceNetStatus, 10)
	sleepTimeSec := 1
	answerTimeoutSec := 1 // Short timeout

	// Start packet handler
	HandlePacket(sleepTimeSec, answerTimeoutSec, serviceConns, statusChan)

	// Wait for timeout status (status should be 0)
	select {
	case status := <-statusChan:
		if status.ServiceName != "test-service" {
			t.Errorf("ServiceName = %v, want test-service", status.ServiceName)
		}
		if status.Status != 0 {
			t.Errorf("Status = %v, want 0 (timeout)", status.Status)
		}
	case <-time.After(3 * time.Second):
		t.Error("Timeout waiting for status update")
	}
}

func TestServiceNetConnection(t *testing.T) {
	conn := ServiceNetConnection{
		ServiceName: "test",
	}

	if conn.ServiceName != "test" {
		t.Errorf("ServiceName = %v, want test", conn.ServiceName)
	}
}
