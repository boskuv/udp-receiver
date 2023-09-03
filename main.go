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

	pc1, err := net.ListenPacket("udp4", ":8829")

	if err != nil {
		panic(err)
	}

	pc2, err := net.ListenPacket("udp4", ":8830")

	if err != nil {
		panic(err)
	}

	defer pc1.Close()
	defer pc2.Close()

	go handlePacket(pc1, "")
	go handlePacket(pc2, "")

	select {}
}

func handlePacket(pc net.PacketConn, port string) {
	buf := make([]byte, 1024)

	lastReceived := time.Now()

	for {
		pc.SetReadDeadline(time.Now().Add(5 * time.Second)) // TODO: cfg parameter
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			fmt.Println("Error reading UDP packet:", err)
			continue
		}

		Prom(lastReceived, port)

		fmt.Printf("%s sent this to port %s: %s\n", addr, port, buf[:n])
		lastReceived = time.Now()
	}
}

func Prom(lastReceived time.Time, port string) {
	if time.Since(lastReceived) > time.Minute {
		caster_status.WithLabelValues(port).Set(0)
	} else {
		caster_status.WithLabelValues(port).Set(1)
	}
	time.Sleep(time.Second)
}
