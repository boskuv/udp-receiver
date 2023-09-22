package services

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

type ServiceNetConnection struct {
	serviceName string
	PacketConn  net.UDPConn
}

func ListenOnPorts(udpServices map[string]int, AnswerTimeoutSec int) (*[]ServiceNetConnection, error) {
	var serviceConns []ServiceNetConnection
	for serviceName, port := range udpServices {

		addr := &net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: port,
		}

		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return nil, fmt.Errorf("ListenOnPorts(..)->net.ListenUDP(..) to service %s: %w", serviceName, err)
		}

		fd, err := conn.File()
		if err != nil {
			return nil, fmt.Errorf("ListenOnPorts(..) Error getting file descriptor in service %s: %w", serviceName, err)
		}

		err = syscall.SetsockoptInt(int(fd.Fd()), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		if err != nil {
			return nil, fmt.Errorf("ListenOnPorts(..) Error setting SO_REUSEPORT option in service %s: %w", serviceName, err)
		}
		fd.Close()

		defer conn.Close()

		conn.SetReadDeadline(time.Now().Add(time.Duration(AnswerTimeoutSec) * time.Second))

		serviceConns = append(serviceConns, ServiceNetConnection{
			serviceName: serviceName,
			PacketConn:  *conn,
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
				_, _, err := conn.PacketConn.ReadFromUDP(dataBuffer)

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
