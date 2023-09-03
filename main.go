package main

import (
	"fmt"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	caster_status = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "job_caster",
			Help: "Location connections number",
		},
		[]string{"caster_status"},
	)
)

func init() {
	prometheus.MustRegister(caster_status)
}

func main() {

	// pc1, err := net.ListenPacket("udp4", ":8829")

	// if err != nil {
	// 	panic(err)
	// }

	// pc2, err := net.ListenPacket("udp4", ":8830")

	// if err != nil {
	// 	panic(err)
	// }

	// defer pc1.Close()
	// defer pc2.Close()

	// go handlePacket(pc1, "")
	// go handlePacket(pc2, "")
	ports := []int{8829, 8830} // add more ports here
	conns, err := listenOnPorts(ports)
	if err != nil {
		panic(err)
	}
	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()
	for i, conn := range conns {
		go handlePacket(conn, fmt.Sprintf("Service %d", i))
	}

	select {} // TODO: what for?
}

func listenOnPorts(ports []int) ([]net.PacketConn, error) {
	var conns []net.PacketConn
	for _, port := range ports {
		addr := fmt.Sprintf(":%d", port)
		conn, err := net.ListenPacket("udp4", addr)
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn) // TODO: slice?
	}
	return conns, nil
}

func handlePacket(pc net.PacketConn, serviceName string) {
	buf := make([]byte, 1024) // TODO: check size

	//lastReceived := time.Now()
	wasPacketReceived := false

	for {
		pc.SetReadDeadline(time.Now().Add(5 * time.Second)) // TODO: cfg parameter
		n, _, err := pc.ReadFrom(buf)
		if err == nil {
			fmt.Printf("UDP packet was received: %s\n", buf[:n])
			wasPacketReceived = true
		}

		Prom(wasPacketReceived, serviceName)

	}
}

func Prom(wasPacketReceived bool, serviceName string) {
	if wasPacketReceived {
		caster_status.WithLabelValues(serviceName).Set(1)
	} else {
		caster_status.WithLabelValues(serviceName).Set(0)
	}
	time.Sleep(time.Second * 5)
}
