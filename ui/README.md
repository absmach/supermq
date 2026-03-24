# FluxMQ Dashboard UI

A Next.js dashboard for real-time monitoring of the FluxMQ message broker. Displays live metrics, session state, cluster topology, and subscription data — all sourced from the FluxMQ Admin API.

## Pages

| Route                      | Description                                                                    |
| -------------------------- | ------------------------------------------------------------------------------ |
| `/dashboard`               | Overview — live metrics, message traffic charts, bandwidth, cluster node table |
| `/dashboard/connections`   | Active connections — connected sessions with protocol, subscriptions, inflight |
| `/dashboard/sessions`      | All sessions — connected and disconnected, filterable, with detail dialog      |
| `/dashboard/subscriptions` | Subscriptions — active topic filters aggregated from connected sessions        |
| `/dashboard/cluster`       | Cluster — node topology with per-node health, sessions, messages, bytes        |
| `/dashboard/broker-info`   | Broker Info — runtime identity, session counts, error counters                 |

## Environment Setup

Copy the example file and edit it:

```bash
cp .env.local.example .env.local
```

Then open `.env.local` and set the values for your environment:

```env
# Primary FluxMQ admin API URL (required to connect to a real broker)
# When unset, the dashboard falls back to mock data.
FLUXMQ_API_URL=http://localhost:9081

# Comma-separated admin API URLs for every node in the cluster.
# Used to fan out per-node stats (sessions, messages, bytes per node).
# For a single-node setup, set this to the same value as FLUXMQ_API_URL.
FLUXMQ_NODE_URLS=http://localhost:9081,http://localhost:9082,http://localhost:9083

# MQTT broker URL for direct browser connections (embedded in the client bundle)
NEXT_PUBLIC_BROKER_URL=mqtt://localhost:1883

# WebSocket URL for MQTT-over-WebSocket connections
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

### Single-node setup

```env
FLUXMQ_API_URL=http://localhost:9081
FLUXMQ_NODE_URLS=http://localhost:9081
```

### 3-node cluster (default ports)

```env
FLUXMQ_API_URL=http://localhost:9081
FLUXMQ_NODE_URLS=http://localhost:9081,http://localhost:9082,http://localhost:9083
```

> `FLUXMQ_API_URL` is the node the dashboard queries for cluster-wide overview data. `FLUXMQ_NODE_URLS` lists all nodes so per-node stats can be fetched in parallel.

## Getting Started

### Prerequisites

- Node.js 18+ or pnpm 10+
- FluxMQ broker running (see `deployments/` for single-node and cluster configs)

### Install dependencies

```bash
pnpm install
```

### Run in development

```bash
pnpm dev
```

Open [http://localhost:3000/dashboard](http://localhost:3000/dashboard).

### Build for production

```bash
pnpm build
pnpm start
```

## Running with Docker

### Build the image

```bash
docker compose -f ui/docker/docker-compose.yaml build
```

### Run standalone (point at an existing broker)

```bash
FLUXMQ_API_URL=http://localhost:8082 \
FLUXMQ_NODE_URLS=http://localhost:8082 \
docker compose -f ui/docker/docker-compose.yaml up
```

### Run with the full single-node stack

Starts FluxMQ and the dashboard together:

```bash
docker compose -f deployments/docker/compose.yaml up -d
```

The dashboard is available at [http://localhost:3000/dashboard](http://localhost:3000/dashboard).

Override the broker URL if needed:

```bash
FLUXMQ_API_URL=http://my-broker:8082 \
docker compose -f deployments/docker/compose.yaml up -d
```

### Run with the Docker cluster

```bash
make docker-cluster-up
docker compose -f deployments/cluster/docker-compose.yaml up -d dashboard
```

The dashboard connects to node 1 (`http://127.0.0.1:9081`) by default and fans out stats across all three nodes.
