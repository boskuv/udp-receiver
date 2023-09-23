package services

import (
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-reuseport"
)

type ServiceNetConnection struct {
	serviceName string
	PacketConn  net.PacketConn
}

func ListenOnPorts(udpServices map[string]int, AnswerTimeoutSec int) (*[]ServiceNetConnection, error) {
	var serviceConns []ServiceNetConnection
	for serviceName, port := range udpServices {

		addr := fmt.Sprintf("0.0.0.0:%d", port)

		conn, err := reuseport.ListenPacket("udp4", addr)
		if err != nil {
			return nil, fmt.Errorf("ListenOnPorts(..)->reuseport.ListenPacket(..) to service %s: %w", serviceName, err)
		}

		conn.SetReadDeadline(time.Now().Add(time.Duration(AnswerTimeoutSec) * time.Second))

		serviceConns = append(serviceConns, ServiceNetConnection{
			serviceName: serviceName,
			PacketConn:  conn,
		})
	}
	return &serviceConns, nil
}

func HandlePacket(SleepTimeSec int, serviceConns *[]ServiceNetConnection, statusChan chan ServiceNetStatus) {
	dataBuffer := make([]byte, 1024)

	for _, sc := range *serviceConns {

		serviceConn := sc

		go func(conn *ServiceNetConnection) {
			var serviceStatus float64

			for {
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
