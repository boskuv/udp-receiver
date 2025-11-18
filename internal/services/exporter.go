package services

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	serviceStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "udp_service_status",
			Help: "UDP service connection status (1 = connected, 0 = disconnected)",
		},
		[]string{"service_name"},
	)
)

// ServiceNetStatus represents the network status of a service
type ServiceNetStatus struct {
	ServiceName string
	Status      float64
}

func init() {
	prometheus.MustRegister(serviceStatusGauge)
}

// ExportToProm exports the service status to Prometheus
func ExportToProm(srv ServiceNetStatus) {
	serviceStatusGauge.WithLabelValues(srv.ServiceName).Set(srv.Status)
}
