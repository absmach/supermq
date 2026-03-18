export interface BrokerStatus {
  node_id: string;
  is_leader: boolean;
  cluster_mode: boolean;
  node_count?: number;
  sessions: number;
  messages_received: number;
  messages_sent: number;
  bytes_received: number;
  bytes_sent: number;
  subscriptions: number;
  retained_messages: number;
  uptime_seconds: number;
  auth_errors: number;
}

export interface NodeInfo {
  node_id: string;
  is_leader: boolean;
  addr: string;
  sessions: number;
  subscriptions: number;
  messages_received: number;
  messages_sent: number;
  bytes_received: number;
  bytes_sent: number;
  uptime_seconds: number;
}

export interface ClientInfo {
  client_id: string;
  connected: boolean;
  subscriptions: number;
  connected_at: string;
  protocol?: string;
  remote?: string;
}

export const MOCK_CLIENTS: ClientInfo[] = [
  { client_id: 'mqtt-client-001',   connected: true,  subscriptions: 5, connected_at: '2026-03-16T07:00:00Z', protocol: 'MQTT', remote: '192.168.1.100:54321' },
  { client_id: 'mqtt-client-002',   connected: true,  subscriptions: 3, connected_at: '2026-03-16T08:30:00Z', protocol: 'MQTT', remote: '192.168.1.101:54322' },
  { client_id: 'mqtt-client-003',   connected: false, subscriptions: 0, connected_at: '2026-03-16T06:00:00Z', protocol: 'MQTT', remote: '192.168.1.102:54323' },
  { client_id: 'sensor-device-001', connected: true,  subscriptions: 1, connected_at: '2026-03-16T09:00:00Z', protocol: 'MQTT', remote: '10.0.0.10:50001' },
  { client_id: 'amqp-conn-1',       connected: true,  subscriptions: 2, connected_at: '2026-03-16T04:00:00Z', protocol: 'AMQP', remote: '10.0.0.50:5672' },
];

const NODE_RATES: Record<string, { msgsIn: number; msgsOut: number; bytesIn: number; bytesOut: number; sessions: number }> = {
  'broker-1': { msgsIn: 12, msgsOut: 11, bytesIn: 4_800, bytesOut: 4_200, sessions: 20 },
  'broker-2': { msgsIn:  7, msgsOut:  6, bytesIn: 2_800, bytesOut: 2_400, sessions: 12 },
  'broker-3': { msgsIn:  5, msgsOut:  4, bytesIn: 2_000, bytesOut: 1_700, sessions: 10 },
};

function jitter(base: number, pct = 0.25): number {
  return Math.round(base * (1 + (Math.random() * 2 - 1) * pct));
}

const _nodes: NodeInfo[] = [
  { node_id: 'broker-1', is_leader: true,  addr: '10.0.0.1:4001', sessions: 20, subscriptions: 87,  messages_received: 58_200,  messages_sent: 56_100,  bytes_received: 23_068_672, bytes_sent: 21_233_664, uptime_seconds: 86_400 },
  { node_id: 'broker-2', is_leader: false, addr: '10.0.0.2:4001', sessions: 12, subscriptions: 62,  messages_received: 41_750,  messages_sent: 40_320,  bytes_received: 17_825_792, bytes_sent: 16_252_928, uptime_seconds: 82_800 },
  { node_id: 'broker-3', is_leader: false, addr: '10.0.0.3:4001', sessions: 10, subscriptions: 38,  messages_received: 28_500,  messages_sent: 27_900,  bytes_received: 11_534_336, bytes_sent: 10_747_904, uptime_seconds: 79_200 },
];

const POLL_S = 5;

function aggregateStatus(): BrokerStatus {
  return {
    node_id: 'broker-1',
    is_leader: true,
    cluster_mode: true,
    node_count: _nodes.length,
    sessions:          _nodes.reduce((s, n) => s + n.sessions, 0),
    subscriptions:     _nodes.reduce((s, n) => s + n.subscriptions, 0),
    messages_received: _nodes.reduce((s, n) => s + n.messages_received, 0),
    messages_sent:     _nodes.reduce((s, n) => s + n.messages_sent, 0),
    bytes_received:    _nodes.reduce((s, n) => s + n.bytes_received, 0),
    bytes_sent:        _nodes.reduce((s, n) => s + n.bytes_sent, 0),
    retained_messages: 34,
    uptime_seconds:    _nodes[0].uptime_seconds,
    auth_errors:       3,
  };
}

export const MOCK_STATUS: BrokerStatus = aggregateStatus();
export const MOCK_NODES: NodeInfo[] = _nodes.map((n) => ({ ...n }));

export interface MockHistoryPoint {
  time: string;
  sessions: number;
  msgsIn: number;
  msgsOut: number;
  bytesIn: number;
  bytesOut: number;
}

export function generateMockHistory(nodeId: string | null, points = 30): MockHistoryPoint[] {
  const rates = nodeId
    ? NODE_RATES[nodeId] ?? NODE_RATES['broker-1']
    : { msgsIn: 24, msgsOut: 21, bytesIn: 9_600, bytesOut: 8_300, sessions: 42 };

  const baseSessions = rates.sessions;
  const history: MockHistoryPoint[] = [];
  const nowMs = Date.now();

  for (let i = points; i >= 1; i--) {
    const t = new Date(nowMs - i * POLL_S * 1000);
    const timeStr = t.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });

    const wave = Math.sin((points - i) / points * Math.PI) * 0.3 + 0.85;
    history.push({
      time:     timeStr,
      sessions: Math.round(baseSessions * (1 + (Math.random() * 0.1 - 0.05))),
      msgsIn:   Math.round(rates.msgsIn   * wave * (1 + (Math.random() * 0.3 - 0.15))),
      msgsOut:  Math.round(rates.msgsOut  * wave * (1 + (Math.random() * 0.3 - 0.15))),
      bytesIn:  Math.round(rates.bytesIn  * wave * (1 + (Math.random() * 0.3 - 0.15))),
      bytesOut: Math.round(rates.bytesOut * wave * (1 + (Math.random() * 0.3 - 0.15))),
    });
  }
  return history;
}

export function formatCount(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

export function formatBytes(bytes: number): string {
  if (bytes >= 1_073_741_824) return `${(bytes / 1_073_741_824).toFixed(1)} GB`;
  if (bytes >= 1_048_576) return `${(bytes / 1_048_576).toFixed(1)} MB`;
  if (bytes >= 1_024) return `${(bytes / 1_024).toFixed(1)} KB`;
  return `${bytes} B`;
}

export function formatUptime(seconds: number): string {
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (d > 0) return `${d}d ${h}h`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}
