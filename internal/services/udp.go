package services

import (
	"fmt"
	"net"
	"time"
)

type ServiceNetConnection struct {
	serviceName string
	PacketConn  net.PacketConn
}

func ListenOnPorts(udpServices map[string]int) ([]ServiceNetConnection, error) {
	var serviceConns []ServiceNetConnection
	for serviceName, port := range udpServices {
		addr := fmt.Sprintf(":%d", port)
		conn, err := net.ListenPacket("udp4", addr)
		if err != nil {
			return nil, err
		}
		serviceConns = append(serviceConns, ServiceNetConnection{
			serviceName: serviceName,
			PacketConn:  conn,
		}) // TODO: slice?
	}
	return serviceConns, nil
}

func HandlePacket(serviceConns []ServiceNetConnection, statusChan chan ServiceNetStatus) {
	buf := make([]byte, 1024) // TODO: check len and cap

	time.Sleep(time.Second) // TODO: cfg parameter

	for _, sc := range serviceConns {

		go func(conn ServiceNetConnection) {
			for {
				conn.PacketConn.SetReadDeadline(time.Now().Add(5 * time.Second))
				n, _, err := conn.PacketConn.ReadFrom(buf)
				if err == nil {
					fmt.Printf("UDP packet was received: %s\n", buf[:n])
					statusChan <- ServiceNetStatus{ServiceName: conn.serviceName, Status: 1}
				} else {
					statusChan <- ServiceNetStatus{ServiceName: conn.serviceName, Status: 0}
				}
			}
		}(sc)
	}

}
