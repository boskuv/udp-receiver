package services

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestExportToProm(t *testing.T) {
	// Reset the metric to avoid interference from other tests
	serviceStatusGauge.Reset()

	tests := []struct {
		name   string
		status ServiceNetStatus
		want   float64
	}{
		{
			name: "status 1",
			status: ServiceNetStatus{
				ServiceName: "test-service",
				Status:      1,
			},
			want: 1,
		},
		{
			name: "status 0",
			status: ServiceNetStatus{
				ServiceName: "test-service",
				Status:      0,
			},
			want: 0,
		},
		{
			name: "different service",
			status: ServiceNetStatus{
				ServiceName: "another-service",
				Status:      1,
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExportToProm(tt.status)

			// Verify the metric value
			metric := &dto.Metric{}
			err := serviceStatusGauge.WithLabelValues(tt.status.ServiceName).Write(metric)
			if err != nil {
				t.Fatalf("Failed to write metric: %v", err)
			}

			if metric.Gauge == nil {
				t.Fatal("Metric gauge is nil")
			}

			if *metric.Gauge.Value != tt.want {
				t.Errorf("ExportToProm() metric value = %v, want %v", *metric.Gauge.Value, tt.want)
			}
		})
	}
}

func TestServiceStatusGauge_Registration(t *testing.T) {
	// Verify that the metric is registered
	registry := prometheus.NewRegistry()
	registry.MustRegister(serviceStatusGauge)

	// Try to register again - should not panic if already registered
	// This test verifies the metric is properly initialized
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	found := false
	for _, mf := range metrics {
		if mf.GetName() == "udp_service_status" {
			found = true
			if mf.GetHelp() != "UDP service connection status (1 = connected, 0 = disconnected)" {
				t.Errorf("Metric help = %v, want 'UDP service connection status (1 = connected, 0 = disconnected)'", mf.GetHelp())
			}
		}
	}

	if !found {
		t.Error("Metric 'udp_service_status' not found in registry")
	}
}
