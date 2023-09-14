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
			return nil, fmt.Errorf("ListenOnPorts(..) to service %s: %w", serviceName, err)
		}
		serviceConns = append(serviceConns, ServiceNetConnection{
			serviceName: serviceName,
			PacketConn:  conn,
		}) // TODO: slice?
	}
	return serviceConns, nil
}

func HandlePacket(SleepTimeSec int, AnswerTimeoutSec int, serviceConns []ServiceNetConnection, statusChan chan ServiceNetStatus) {
	dataBuffer := make([]byte, 1024)

	for _, sc := range serviceConns {

		serviceConn := sc // TODO: check

		go func(conn *ServiceNetConnection) {
			var serviceStatus float64

			for {
				conn.PacketConn.SetReadDeadline(time.Now().Add(time.Duration(AnswerTimeoutSec) * time.Second))
				_, _, err := conn.PacketConn.ReadFrom(dataBuffer)

				serviceStatus = 0

				if err == nil {
					serviceStatus = 1
				}

				statusChan <- ServiceNetStatus{ServiceName: conn.serviceName, Status: serviceStatus}

				time.Sleep(time.Duration(SleepTimeSec) * time.Second)
			}
		}(&serviceConn)
	}

}
