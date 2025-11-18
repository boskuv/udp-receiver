package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boskuv/udp-receiver/internal/config"
	"github.com/boskuv/udp-receiver/internal/services"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Run starts the UDP receiver application
func Run(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start UDP listeners
	serviceConns, err := services.ListenOnPorts(cfg.Services, cfg.AnswerTimeoutSec)
	if err != nil {
		return fmt.Errorf("failed to start UDP listeners: %w", err)
	}

	// Ensure connections are closed on exit
	defer func() {
		log.Println("Closing UDP connections...")
		for _, sc := range serviceConns {
			if err := sc.PacketConn.Close(); err != nil {
				log.Printf("Error closing connection for service %s: %v", sc.ServiceName, err)
			}
		}
	}()

	// Start Prometheus metrics server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", healthCheckHandler)

	server := &http.Server{
		Addr:    cfg.PromAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting Prometheus metrics server on %s", cfg.PromAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting metrics server: %v", err)
		}
	}()

	// Start packet handlers
	statusChan := make(chan services.ServiceNetStatus, len(cfg.Services)*2)
	services.HandlePacket(cfg.SleepTimeSec, cfg.AnswerTimeoutSec, serviceConns, statusChan)

	// Start status processor
	go processStatusUpdates(ctx, statusChan)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	cancel()
	return nil
}

// processStatusUpdates processes status updates from UDP handlers
func processStatusUpdates(ctx context.Context, statusChan <-chan services.ServiceNetStatus) {
	for {
		select {
		case <-ctx.Done():
			return
		case currentServiceStatus := <-statusChan:
			log.Printf("Service '%s' status: %.0f", currentServiceStatus.ServiceName, currentServiceStatus.Status)
			services.ExportToProm(currentServiceStatus)
		}
	}
}

// healthCheckHandler provides a simple health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
