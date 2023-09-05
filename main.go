package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type ServiceNetConnection struct {
	serviceName string
	pc          net.PacketConn
}

var (
	caster_status = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "job_caster",
			Help: "Location connections number",
		},
		[]string{"caster_status"},
	)
)

var mutex sync.Mutex

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

	for {
		scChan := make(chan ServiceNetConnection)
		for _, sc := range serviceConns {
			scChan <- sc
			go handlePacket(scChan)
		}

		select {} // TODO: what for?
	}
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

func handlePacket(scChan chan ServiceNetConnection) { //sc ServiceNetConnection) {
	buf := make([]byte, 1024) // TODO: check size

	//var sc ServiceNetConnection
	sc := <-scChan
	wasPacketReceived := false

	//for {
	sc.pc.SetReadDeadline(time.Now().Add(5 * time.Second)) // TODO: cfg parameter
	n, _, err := sc.pc.ReadFrom(buf)
	if err == nil {
		fmt.Printf("UDP packet was received: %s\n", buf[:n])
		wasPacketReceived = true
	}

	Prom(wasPacketReceived, sc.serviceName)

	//}
}

func Prom(wasPacketReceived bool, serviceName string) {

	mutex.Lock()

	if wasPacketReceived {
		caster_status.WithLabelValues(serviceName).Set(1)
	} else {
		caster_status.WithLabelValues(serviceName).Set(0)
	}

	defer mutex.Unlock()

	time.Sleep(time.Second * 1)
}
