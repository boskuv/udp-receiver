package services

import "github.com/prometheus/client_golang/prometheus"

var (
	caster_status = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "job_caster",
			Help: "Location connections number",
		},
		[]string{"caster_status"},
	)
)

type ServiceNetStatus struct {
	ServiceName string
	Status      float64
}

func init() {
	prometheus.MustRegister(caster_status)
}

func ExportToProm(srv ServiceNetStatus) {
	caster_status.WithLabelValues(srv.ServiceName).Set(srv.Status)
}
