# UDP Receiver

A Go application that monitors UDP services by listening on configured ports and exports metrics to Prometheus.

## Features

- **Multi-service UDP monitoring**: Listen on multiple UDP ports simultaneously
- **Prometheus metrics**: Export service status metrics for monitoring
- **Graceful shutdown**: Proper cleanup on termination
- **Health check endpoint**: HTTP endpoint for health checks
- **Configurable**: YAML-based configuration

## Requirements

- Go 1.21 or higher
- Linux (for UDP port reuse support)

## Installation

### From source

```bash
git clone <repository-url>
cd udp-receiver
make deps
make build
```

### Using Docker

```bash
docker build -t udp-receiver .
```

## Configuration

Create a `config.yml` file based on `config.yml.example`:

```yaml
promAddr: :8080              # Prometheus metrics endpoint address
services:                    # Map of service names to UDP ports
  first: 8830
  second: 8831
sleepTimeSec: 1              # Sleep time between packet reads (seconds)
answerTimeoutSec: 5          # Timeout for reading packets (seconds)
```

### Configuration Options

- `promAddr`: Address for the Prometheus metrics HTTP server (e.g., `:8080`, `0.0.0.0:8080`)
- `services`: Map of service names to UDP port numbers
- `sleepTimeSec`: Time to wait between packet read attempts
- `answerTimeoutSec`: Maximum time to wait for a packet before timing out

## Usage

### Running locally

```bash
./bin/ussc -config ./cmd/ussc/config.yml
```

Or using Make:

```bash
make run
```

### Running with Docker

```bash
docker run --rm \
  -p 8080:8080 \
  -p 8830:8830/udp \
  -p 8831:8831/udp \
  -v $(pwd)/cmd/ussc/config.yml:/app/config.yml \
  udp-receiver
```

Or using Make:

```bash
make docker
```

## Metrics

The application exposes Prometheus metrics at `/metrics`:

- `udp_service_status{service_name="<name>"}`: Service status (1 = packet received, 0 = timeout/no packet)

### Example Query

```promql
# Average status over 5 minutes
avg_over_time(udp_service_status[5m])
```

## Endpoints

- `GET /metrics`: Prometheus metrics endpoint
- `GET /health`: Health check endpoint (returns 200 OK)

## Development

### Running tests

```bash
make test
```

### Code formatting

```bash
make fmt
```

### Linting

```bash
make lint
```

### Building

```bash
make build
```

## Project Structure

```
.
├── cmd/
│   └── ussc/              # Main application entry point
│       ├── main.go
│       └── config.yml.example
├── internal/
│   ├── config/            # Configuration management
│   ├── services/          # UDP service handlers and Prometheus exporter
│   └── ussc/              # Application logic
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## How It Works

1. The application reads configuration from a YAML file
2. For each configured service, it starts a UDP listener on the specified port
3. Each service runs in a separate goroutine that:
   - Waits for UDP packets with a configurable timeout
   - Sends status updates (1 = packet received, 0 = timeout) to a channel
4. A status processor reads from the channel and exports metrics to Prometheus
5. An HTTP server exposes the Prometheus metrics endpoint

## Troubleshooting

### Port already in use

If you get an error about a port being in use, either:
- Change the port in your configuration
- Stop the process using that port
- Use `SO_REUSEPORT` (already enabled via `go-reuseport`)

### No metrics appearing

- Check that the Prometheus metrics endpoint is accessible: `curl http://localhost:8080/metrics`
- Verify that UDP packets are being sent to the configured ports
- Check application logs for errors

