package services

import (
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-reuseport"
)

const (
	// DefaultBufferSize is the default buffer size for reading UDP packets
	DefaultBufferSize = 1024
)

// ServiceNetConnection represents a UDP connection for a service
type ServiceNetConnection struct {
	ServiceName string
	PacketConn  net.PacketConn
}

// ListenOnPorts creates UDP listeners for all configured services
func ListenOnPorts(udpServices map[string]int, answerTimeoutSec int) ([]ServiceNetConnection, error) {
	if len(udpServices) == 0 {
		return nil, fmt.Errorf("no services configured")
	}

	serviceConns := make([]ServiceNetConnection, 0, len(udpServices))
	for serviceName, port := range udpServices {
		if port <= 0 || port > 65535 {
			return nil, fmt.Errorf("invalid port %d for service %s", port, serviceName)
		}

		addr := fmt.Sprintf("0.0.0.0:%d", port)

		conn, err := reuseport.ListenPacket("udp4", addr)
		if err != nil {
			// Close already opened connections on error
			for _, sc := range serviceConns {
				sc.PacketConn.Close()
			}
			return nil, fmt.Errorf("failed to listen on port %d for service %s: %w", port, serviceName, err)
		}

		serviceConns = append(serviceConns, ServiceNetConnection{
			ServiceName: serviceName,
			PacketConn:  conn,
		})
	}
	return serviceConns, nil
}

// HandlePacket starts goroutines to handle UDP packets for each service
func HandlePacket(sleepTimeSec int, answerTimeoutSec int, serviceConns []ServiceNetConnection, statusChan chan<- ServiceNetStatus) {
	dataBuffer := make([]byte, DefaultBufferSize)

	for i := range serviceConns {
		serviceConn := &serviceConns[i]

		go func(conn *ServiceNetConnection) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic and send error status
					statusChan <- ServiceNetStatus{
						ServiceName: conn.ServiceName,
						Status:      0,
					}
				}
			}()

			for {
				timeout := time.Duration(answerTimeoutSec) * time.Second
				conn.PacketConn.SetReadDeadline(time.Now().Add(timeout))

				_, _, err := conn.PacketConn.ReadFrom(dataBuffer)

				serviceStatus := float64(0)
				if err == nil {
					serviceStatus = 1
				}

				select {
				case statusChan <- ServiceNetStatus{
					ServiceName: conn.ServiceName,
					Status:      serviceStatus,
				}:
				default:
					// Channel is full, skip this status update
				}

				time.Sleep(time.Duration(sleepTimeSec) * time.Second)
			}
		}(serviceConn)
	}
}
