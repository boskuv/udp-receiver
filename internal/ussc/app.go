package app

import (
	"fmt"
	"net/http"
	"time"

	"udp-receiver/internal/config"
	"udp-receiver/internal/services"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run(cfg *config.Config) { // TODO: pointer?
	udpServices := cfg.Services

	serviceConns, err := services.ListenOnPorts(udpServices)
	if err != nil {
		panic(err)
	}

	defer func() {
		for _, sc := range serviceConns {
			sc.PacketConn.Close()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	//addr := "localhost:8080"
	//log.Printf("Starting web server at %s\n", *addr)

	go func() {
		err := http.ListenAndServe(cfg.PromAddr, nil)
		if err != nil {
			panic(err)
		}
	}()

	statusChan := make(chan services.ServiceNetStatus) // TODO: check cap and len (for example 2)
	services.HandlePacket(cfg.SleepTimeSec, cfg.AnswerTimeoutSec, serviceConns, statusChan)

	for {
		select {
		case currentServiceStatus := <-statusChan:
			fmt.Println(time.Now().String(), " | Received result:", currentServiceStatus, " | chan cap/len: ", cap(statusChan), "/", len(statusChan))
			services.ExportToProm(currentServiceStatus)
		default:
			time.Sleep(time.Second)
		}
	}

}
