package main

import (
	"fmt"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type ServiceNetConnection struct {
	serviceName string
	pc          net.PacketConn
}

type ServiceNetStatus struct {
	serviceName string
	status      float64
}

var (
	caster_status = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "job_caster",
			Help: "Location connections number",
		},
		[]string{"caster_status"}, // TODO: rename
	)
)

func init() {
	prometheus.MustRegister(caster_status)
}

func main() {
	//ports := []int{8829, 8830} // add more ports here
	udpServices := make(map[string]int)
	udpServices["first"] = 8829
	udpServices["second"] = 8830

	serviceConns, err := listenOnPorts(udpServices)
	if err != nil {
		panic(err)
	}

	defer func() {
		for _, sc := range serviceConns {
			sc.pc.Close()
		}
	}()

	statusChan := make(chan ServiceNetStatus, 100) // check cap and len
	go handlePacket(serviceConns, statusChan)

	for {
		select {
		case currentServiceStatus := <-statusChan:
			fmt.Println(time.Now().String(), " | Received result:", currentServiceStatus)
			ExportToProm(currentServiceStatus)
		default:
			fmt.Println("No result yet, continuing...")
			time.Sleep(time.Second)
		}
	}

	//select {} // TODO: what for?
}

func listenOnPorts(udpServices map[string]int) ([]ServiceNetConnection, error) {
	var serviceConns []ServiceNetConnection
	for serviceName, port := range udpServices {
		addr := fmt.Sprintf(":%d", port)
		conn, err := net.ListenPacket("udp4", addr)
		if err != nil {
			return nil, err
		}
		serviceConns = append(serviceConns, ServiceNetConnection{
			serviceName: serviceName,
			pc:          conn,
		}) // TODO: slice?
	}
	return serviceConns, nil
}

func handlePacket(serviceConns []ServiceNetConnection, statusChan chan ServiceNetStatus) {
	buf := make([]byte, 1024) // TODO: check len and cap

	// TODO: sleep
	for {

		for _, sc := range serviceConns {

			sc.pc.SetReadDeadline(time.Now().Add(5 * time.Second)) // TODO: cfg parameter
			n, _, err := sc.pc.ReadFrom(buf)

			if err == nil {
				fmt.Printf("UDP packet was received: %s\n", buf[:n]) // TODO: remove
				statusChan <- ServiceNetStatus{serviceName: sc.serviceName, status: 1}
				fmt.Println(time.Now().String(), "serviceName: ", sc.serviceName, "status: ", 0)
				continue
			}
			// TODO: else
			statusChan <- ServiceNetStatus{serviceName: sc.serviceName, status: 0}

		}

	}
}

func ExportToProm(srv ServiceNetStatus) {
	caster_status.WithLabelValues(srv.serviceName).Set(srv.status)
}
