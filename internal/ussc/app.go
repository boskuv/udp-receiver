package app

import (
	"fmt"
	"time"

	"udp-receiver/internal/services"
)

func Run() {
	//ports := []int{8829, 8830} // add more ports here
	udpServices := make(map[string]int)
	udpServices["first"] = 8829
	udpServices["second"] = 8830

	serviceConns, err := services.ListenOnPorts(udpServices)
	if err != nil {
		panic(err)
	}

	defer func() {
		for _, sc := range serviceConns {
			sc.PacketConn.Close()
		}
	}()

	statusChan := make(chan services.ServiceNetStatus) // check cap and len, 2
	services.HandlePacket(serviceConns, statusChan)

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
