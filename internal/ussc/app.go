package app

import (
	"fmt"
	"net/http"
	"time"

	"udp-receiver/internal/config"
	"udp-receiver/internal/services"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run(cfg *config.Config) {
	udpServices := cfg.Services

	serviceConns, err := services.ListenOnPorts(udpServices, cfg.AnswerTimeoutSec)
	if err != nil {
		panic(err)
	}

	defer func() {
		for _, sc := range *serviceConns {
			sc.PacketConn.Close()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		err := http.ListenAndServe(cfg.PromAddr, nil)
		if err != nil {
			panic(err)
		}
	}()

	statusChan := make(chan services.ServiceNetStatus, len(cfg.Services))
	services.HandlePacket(cfg.SleepTimeSec, serviceConns, statusChan)

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
