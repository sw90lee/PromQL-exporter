# PromQL Exporter

A Go-based Prometheus metrics exporter for 5G/CNF (Cloud Native Functions) monitoring. This exporter collects metrics from various sources including CSV files, Kubernetes clusters, and HTTP endpoints, exposing them in Prometheus format.

## Overview

This project is a specialized Prometheus exporter designed for 5G Core Network Functions (CNF) monitoring, particularly focused on Samsung CPC (Control Plane Component) metrics. It collects and exports metrics from multiple sources:

- CSV-based metrics from OSS (Operation Support System)
- Kubernetes cluster metrics via API
- HTTP endpoint scraping
- Real-time performance data aggregation

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   OSS System    │    │  Kubernetes API  │    │  HTTP Endpoints │
│   (CSV Files)   │    │                  │    │                 │
└─────────┬───────┘    └────────┬─────────┘    └─────────┬───────┘
          │                     │                        │
          └─────────────────────┼────────────────────────┘
                                │
                    ┌───────────▼──────────┐
                    │  PromQL Exporter     │
                    │  - Data Collection   │
                    │  - Metric Processing │
                    │  - File Management   │
                    └───────────┬──────────┘
                                │
                    ┌───────────▼──────────┐
                    │  Prometheus Format   │
                    │  HTTP Endpoints      │
                    │  /metrics, /api/...  │
                    └──────────────────────┘
```

## Features

- **Multi-source Data Collection**: Supports CSV files, Kubernetes APIs, and HTTP endpoints
- **Automated File Management**: Automatic backup and organization of collected data
- **5G/CNF Metrics**: Specialized metrics for 5G Core Network Functions
- **Kubernetes Integration**: Native support for OpenShift/Kubernetes cluster monitoring
- **RESTful API**: Provides API endpoints for metric access
- **Containerized Deployment**: Docker support with configurable parameters

## Project Structure

```
PromQL-exporter/
├── cmd/
│   └── exporter.go          # Main application entry point
├── pkg/
│   ├── exporter/            # Core exporter logic
│   │   ├── appCollector.go  # Application metrics collector
│   │   └── model.go         # Data models and structures
│   ├── csv/                 # CSV file handling
│   ├── curl/                # HTTP client utilities
│   ├── k8sClient/           # Kubernetes client
│   ├── metricApi/           # API handlers
│   └── utils/               # Utility functions
├── cfg/                     # Configuration management
├── logger/                  # Logging utilities
├── config.yml               # Main configuration
├── app_config.yml           # Application-specific metrics config
├── cnf_config.yml           # CNF metrics configuration
├── Dockerfile               # Container build configuration
└── Makefile                 # Build automation
```

## Configuration

### Main Configuration (`config.yml`)

```yaml
logging:
  LEVEL: INFO
  ENCODE: json

file:
  MEC_CONFIG: "/mnt/data/config"
  CSV_PATH: "/mnt/data/exporter"
  API_PATH: "/mnt/data/exporter/api"
  FAMILY_NAME: ["UECON_AMF", "TMSI_AMF", ...]

exporter:
  CURL_URL: "https://your-oss-system/oss/performanceData"
  OSS_USERNAME: "username"
  OSS_PASSWORD: "password"
```

### Application Metrics (`app_config.yml`)

Defines metrics collection from various endpoints:

```yaml
cpu/metrics:
  collects:
  - metrics:
      node_cpu_seconds_total:
        type: counter
        prefix: p5g_wrcp1
        description: wrcp1_core_cpu_value
        url: "http://prometheus-api/api/v1/query?query=node_cpu_seconds_total"
```

### CNF Metrics (`cnf_config.yml`)

Configures 5G CNF-specific metrics:

```yaml
metrics:
  amf_ue_connect_attempt_count:
    type: counter
    description: "UECON_AMF"
    labels: ["ne_id", "system_id", "ne_name", ...]
    value_sequence: 7
```

## Installation & Deployment

### Docker Build

```bash
# Build image
make save

# Or build manually
docker build -t promql-exporter:latest .
```

### Local Development

```bash
# Install dependencies
go mod download

# Build application
go build -o bin/exporter ./cmd/exporter.go

# Run with configuration
./bin/exporter -metricConfig cnf_config.yml -config-metrics app_config.yml
```

### Kubernetes Deployment

```bash
# Deploy using Helm
helm install promql-exporter ./chart --values values.yaml
```

## Usage

### Running the Exporter

```bash
# Default configuration
./exporter

# Custom configuration files
./exporter -metricConfig custom_cnf.yml -config-metrics custom_app.yml
```

### API Endpoints

- `GET /metrics` - Prometheus metrics endpoint
- `GET /api/metrics` - CNF metrics API
- `GET /cpu/metrics` - CPU metrics endpoint
- `GET /mem/metrics` - Memory metrics endpoint
- `GET /pod/metrics` - Pod metrics endpoint

### Sample Metrics Output

```prometheus
# HELP p5g_exporter_amf_ue_connect_attempt_count AMF UE connection attempts
# TYPE p5g_exporter_amf_ue_connect_attempt_count counter
p5g_exporter_amf_ue_connect_attempt_count{ne_id="amf-001",location="datacenter-1"} 1500

# HELP p5g_mec_node_cpu_seconds_total MEC CPU seconds total
# TYPE p5g_mec_node_cpu_seconds_total counter
p5g_mec_node_cpu_seconds_total{container="prometheus",cpu="0",instance="node-1"} 12345.67
```

## Monitored Metrics Categories

### 5G Core Network Functions
- **AMF (Access and Mobility Management Function)**
  - UE connection attempts/success
  - TMSI (Temporary Mobile Subscriber Identity) operations
  - Transaction processing statistics

- **SMF (Session Management Function)**
  - GTP-C tunnel endpoint operations
  - Session establishment metrics
  - Transaction throughput

- **UPF (User Plane Function)**
  - Packet inspection statistics
  - Data forwarding traffic
  - PDCP volume and packet metrics

### Infrastructure Metrics
- **CPU Utilization**: Node-level CPU metrics across clusters
- **Memory Usage**: Memory consumption and availability
- **Pod Metrics**: Container resource utilization
- **Network Metrics**: Air interface MAC packet statistics

## Data Flow

1. **Collection Phase**:
   - Fetch performance data from OSS via HTTP
   - Query Kubernetes APIs for infrastructure metrics
   - Process CSV files for historical data

2. **Processing Phase**:
   - Parse and validate metric data
   - Apply metric type conversions
   - Generate Prometheus-compatible labels

3. **Export Phase**:
   - Expose metrics via HTTP endpoints
   - Backup processed files
   - Maintain data retention policies

## Development

### Prerequisites

- Go 1.20+
- Docker
- Kubernetes cluster access (optional)
- Access to OSS system for CSV data

### Building

```bash
# Cross-compilation for different platforms
make cc        # Windows
make cclinux   # Linux ARM/ARM64

# Docker image with versioning
make save tag=v1.0.0
```

### Adding New Metrics

1. Define metric in `cnf_config.yml`:
```yaml
new_metric_name:
  type: gauge
  description: "FAMILY_NAME"
  labels: ["label1", "label2"]
  value_sequence: 10
```

2. Add family name to `config.yml`:
```yaml
FAMILY_NAME: [..., "NEW_FAMILY_NAME"]
```

3. Implement collection logic in appropriate collector

## Troubleshooting

### Common Issues

1. **CSV File Not Found**
   - Check file paths in configuration
   - Verify OSS system connectivity
   - Ensure proper file permissions

2. **Kubernetes Authentication**
   - Verify service account permissions
   - Check cluster connectivity
   - Validate token generation

3. **Memory Issues**
   - Monitor large CSV file processing
   - Adjust backup intervals
   - Check disk space for data paths

### Logs

Application uses structured JSON logging:

```bash
# View logs in container
docker logs promql-exporter

# Filter by log level
docker logs promql-exporter 2>&1 | grep ERROR
```

## Security Considerations

- Sensitive URLs and credentials are masked in configuration examples
- Uses bearer token authentication for Kubernetes API access
- TLS verification can be configured for HTTP clients
- File permissions are set appropriately for data directories

## Performance

- Processes metrics in configurable intervals
- Implements file-based caching for large datasets
- Supports concurrent metric collection
- Optimized for containerized deployment

## License

This software is proprietary to KT Corp. All rights reserved. Usage is restricted under license agreement.

---

**Note**: This exporter is specifically designed for Samsung CPC 5G network monitoring and may require customization for other environments.