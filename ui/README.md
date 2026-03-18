# FluxMQ Dashboard UI

A modern Next.js dashboard for real-time monitoring and management of the FluxMQ message broker. Inspired by MonsterMQ, featuring a dark theme design, real-time metrics, and comprehensive broker management capabilities.

## Features

- **Real-time Dashboard**: Monitor broker metrics, connections, and message traffic
- **Client Management**: View and manage connected MQTT clients
- **Topic Browser**: Browse and monitor topics with detailed statistics
- **Connection Monitoring**: Track all broker connections and their stats
- **Storage Metrics**: Monitor database storage, message queues, and disk space
- **Configuration Management**: View and edit broker configurations
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Dark Theme**: Modern dark UI inspired by MonsterMQ

## Pages

- `/dashboard` - Main overview with real-time metrics and charts
- `/dashboard/clients` - MQTT client management
- `/dashboard/topics` - Topic browser and management
- `/dashboard/connections` - Connection monitoring
- `/dashboard/storage` - Storage and database metrics
- `/dashboard/config` - Broker configuration

## Getting Started

### Prerequisites

- Node.js 18+ or pnpm 10+
- FluxMQ broker running (for API integration)

### Installation

```bash
# Install dependencies
npm install
# or
pnpm install
```

### Development

```bash
# Start development server
npm run dev
# or
pnpm dev
```

Open [http://localhost:3000/dashboard](http://localhost:3000/dashboard) to view it in your browser.

### Build

```bash
# Build for production
npm run build

# Start production server
npm start
```
