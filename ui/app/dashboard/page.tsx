'use client';

import { useEffect, useRef, useState } from 'react';
import {
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  LineChart, Line, Legend,
} from 'recharts';
import {
  Activity, Wifi, WifiOff, Server, MessageSquare, HardDrive,
  BookOpen, Clock, Pause, Play, Network,
} from 'lucide-react';
import { useTheme } from '@/lib/theme-provider';
import {
  BrokerStatus, NodeInfo,
  formatCount, formatBytes, formatUptime,
  MOCK_STATUS, MOCK_NODES,
  generateMockHistory,
} from '@/lib/api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

const POLL_MS = 5_000;
const POLL_S = POLL_MS / 1000;
const MAX_POINTS = 40;

interface ChartPoint {
  time: string;
  sessions: number;
  msgsIn: number;
  msgsOut: number;
  bytesIn: number;
  bytesOut: number;
}

function now(): string {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

function nodeToStatus(node: NodeInfo): BrokerStatus {
  return {
    node_id: node.node_id,
    is_leader: node.is_leader,
    cluster_mode: false,
    sessions: node.sessions,
    messages_received: node.messages_received,
    messages_sent: node.messages_sent,
    bytes_received: node.bytes_received,
    bytes_sent: node.bytes_sent,
    subscriptions: node.subscriptions,
    retained_messages: 0,
    uptime_seconds: node.uptime_seconds,
    auth_errors: 0,
  };
}


export default function Dashboard() {
  const { theme } = useTheme();
  const isDark = theme === 'dark';
  const [clusterStatus, setClusterStatus] = useState<BrokerStatus>(MOCK_STATUS);
  const [nodes, setNodes] = useState<NodeInfo[]>(MOCK_NODES);
  const initialHistories = (): Record<string, ChartPoint[]> => {
    const init: Record<string, ChartPoint[]> = { '': generateMockHistory(null) };
    for (const node of MOCK_NODES) {
      init[node.node_id] = generateMockHistory(node.node_id);
    }
    return init;
  };
  const historiesRef = useRef<Record<string, ChartPoint[]>>(initialHistories());
  const [, forceRender] = useState(0);
  const prevRef = useRef<Record<string, { msgsRx: number; msgsTx: number; bytesRx: number; bytesTx: number }>>({
    '': { msgsRx: MOCK_STATUS.messages_received, msgsTx: MOCK_STATUS.messages_sent, bytesRx: MOCK_STATUS.bytes_received, bytesTx: MOCK_STATUS.bytes_sent },
  });
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [liveUpdates, setLiveUpdates] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function poll() {
      if (!liveUpdates) return;
      try {
        if (cancelled) return;

        setClusterStatus(MOCK_STATUS);
        setNodes(MOCK_NODES);
        setError(null);

        const ts = now();
        const histories = historiesRef.current;
        const prev = prevRef.current;
        function makePoint(
          sessions: number,
          msgsRx: number, msgsTx: number,
          bytesRx: number, bytesTx: number,
          key: string,
        ): ChartPoint {
          const p = prev[key] ?? { msgsRx: 0, msgsTx: 0, bytesRx: 0, bytesTx: 0 };
          const point: ChartPoint = {
            time: ts,
            sessions,
            msgsIn: Math.max(0, (msgsRx - p.msgsRx) / POLL_S),
            msgsOut: Math.max(0, (msgsTx - p.msgsTx) / POLL_S),
            bytesIn: Math.max(0, (bytesRx - p.bytesRx) / POLL_S),
            bytesOut: Math.max(0, (bytesTx - p.bytesTx) / POLL_S),
          };
          prev[key] = { msgsRx, msgsTx, bytesRx, bytesTx };
          return point;
        }

        histories[''] = [
          ...(histories[''] ?? []),
          makePoint(MOCK_STATUS.sessions, MOCK_STATUS.messages_received, MOCK_STATUS.messages_sent, MOCK_STATUS.bytes_received, MOCK_STATUS.bytes_sent, ''),
        ].slice(-MAX_POINTS);

        for (const node of MOCK_NODES) {
          histories[node.node_id] = [
            ...(histories[node.node_id] ?? []),
            makePoint(node.sessions, node.messages_received, node.messages_sent, node.bytes_received, node.bytes_sent, node.node_id),
          ].slice(-MAX_POINTS);
        }

        forceRender((n) => n + 1);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : 'Failed to reach broker');
      } finally {
        // no-op — loading state removed, pre-seeded history shown immediately
      }
    }

    poll();
    const id = setInterval(poll, POLL_MS);
    return () => { cancelled = true; clearInterval(id); };
  }, [liveUpdates]);

  const displayStatus: BrokerStatus =
    selectedNodeId === null
      ? clusterStatus
      : nodeToStatus(nodes.find((n) => n.node_id === selectedNodeId) ?? nodes[0]);

  const historyKey = selectedNodeId ?? '';
  const displayHistory = historiesRef.current[historyKey] ?? [];

  const gridColor = isDark ? '#334155' : '#e2e8f0';
  const axisColor = isDark ? '#94a3b8' : '#64748b';
  const tooltipBg = isDark ? '#1e293b' : '#ffffff';
  const tooltipBdr = isDark ? '#475569' : '#e2e8f0';
  const tooltipTxt = isDark ? '#f1f5f9' : '#0f172a';

  const tooltipStyle = {
    backgroundColor: tooltipBg,
    border: `1px solid ${tooltipBdr}`,
    borderRadius: '8px',
    color: tooltipTxt,
    fontSize: 12,
  };

  const statCards = [
    { label: 'Active Connections', value: formatCount(displayStatus.sessions), icon: Activity, gradient: 'from-flux-blue-600 to-flux-blue-400' },
    { label: 'Subscriptions', value: formatCount(displayStatus.subscriptions), icon: BookOpen, gradient: 'from-flux-purple to-flux-dark' },
    { label: 'Messages Received', value: formatCount(displayStatus.messages_received), icon: MessageSquare, gradient: 'from-flux-green to-flux-teal' },
    { label: 'Messages Sent', value: formatCount(displayStatus.messages_sent), icon: MessageSquare, gradient: 'from-flux-teal to-flux-blue-400' },
    { label: 'Bytes In', value: formatBytes(displayStatus.bytes_received), icon: HardDrive, gradient: 'from-flux-orange to-amber-400' },
    { label: 'Bytes Out', value: formatBytes(displayStatus.bytes_sent), icon: HardDrive, gradient: 'from-amber-400 to-flux-orange' },
    { label: 'Retained Messages', value: formatCount(displayStatus.retained_messages), icon: Server, gradient: 'from-flux-purple to-flux-teal' },
    { label: 'Uptime', value: formatUptime(displayStatus.uptime_seconds), icon: Clock, gradient: 'from-flux-green to-flux-blue-400' },
  ];

  const scopeLabel = selectedNodeId ?? 'Cluster';

  return (
    <div className="p-6 lg:p-8 space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-flux-text">Dashboard</h1>
          <p className="text-flux-text-muted mt-1 text-sm">
            {clusterStatus.cluster_mode
              ? <>Cluster · <span className="font-medium text-flux-text">{nodes.length} nodes</span> · via <span className="font-medium text-flux-text">{clusterStatus.node_id}</span></>
              : <>Single node · <span className="font-medium text-flux-text">{clusterStatus.node_id}</span></>
            }
          </p>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setLiveUpdates((v) => !v)}
            className={`flex items-center gap-1.5 text-xs ${liveUpdates
                ? 'border-flux-green/40 text-flux-green bg-flux-green/10 hover:bg-flux-green/20'
                : 'border-flux-card-border text-flux-text-muted hover:bg-flux-hover'
              }`}
          >
            {liveUpdates ? <><Pause className="w-3 h-3" /> Live</> : <><Play className="w-3 h-3" /> Paused</>}
          </Button>

          <Badge
            variant="outline"
            className={`flex items-center gap-1.5 text-sm px-3 py-1.5 ${error
                ? 'text-flux-red border-flux-red/30 bg-flux-red/10'
                : 'text-flux-green border-flux-green/30 bg-flux-green/10'
              }`}
          >
            {error ? <WifiOff className="w-4 h-4" /> : <Wifi className="w-4 h-4" />}
            {error ? 'Offline' : 'Online'}
          </Badge>
        </div>
      </div>
      {error && (
        <div className="rounded-lg border border-flux-red/30 bg-flux-red/10 px-4 py-3 text-sm text-flux-red">
          ⚠ {error} — showing last known data, retrying every {POLL_MS / 1000}s
        </div>
      )}
      <div className="flex items-center gap-2 flex-wrap">
        <Network className="w-4 h-4 text-flux-text-muted shrink-0" />
        <NodePill label="Cluster" active={selectedNodeId === null} onClick={() => setSelectedNodeId(null)} />
        {nodes.map((node) => (
          <NodePill
            key={node.node_id}
            label={node.node_id}
            active={selectedNodeId === node.node_id}
            isLeader={node.is_leader}
            onClick={() => setSelectedNodeId(node.node_id)}
          />
        ))}
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        {statCards.map((card) => {
          const Icon = card.icon;
          return (
            <Card key={card.label} className="border-flux-card-border bg-flux-card">
              <CardContent className="p-5">
                <div className={`p-2 rounded-lg bg-gradient-to-br ${card.gradient} w-fit mb-3`}>
                  <Icon className="w-4 h-4 text-white" />
                </div>
                <p className="text-flux-text-muted text-xs mb-1">{card.label}</p>
                <p className="text-2xl font-bold text-flux-text">{card.value}</p>
              </CardContent>
            </Card>
          );
        })}
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="border-flux-card-border bg-flux-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-base text-flux-text">Message Traffic Trends</CardTitle>
            <p className="text-xs text-flux-text-muted">
              {scopeLabel} · msgs/s in &amp; out · polled every {POLL_S}s
            </p>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={260}>
              <LineChart data={displayHistory} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
                <XAxis dataKey="time" stroke={axisColor} tick={{ fontSize: 10 }} interval="preserveStartEnd" />
                <YAxis stroke={axisColor} tick={{ fontSize: 10 }} allowDecimals={false} unit="/s" />
                <Tooltip contentStyle={tooltipStyle} formatter={(v: number) => [`${v.toFixed(1)}/s`]} />
                <Legend wrapperStyle={{ fontSize: 11 }} />
                <Line type="monotone" dataKey="msgsIn" name="In" stroke="#10b981" strokeWidth={2} dot={false} isAnimationActive={false} />
                <Line type="monotone" dataKey="msgsOut" name="Out" stroke="#3b82f6" strokeWidth={2} dot={false} isAnimationActive={false} />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card className="border-flux-card-border bg-flux-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-base text-flux-text">Bandwidth</CardTitle>
            <p className="text-xs text-flux-text-muted">
              {scopeLabel} · bytes/s in &amp; out
            </p>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={260}>
              <LineChart data={displayHistory} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
                <XAxis dataKey="time" stroke={axisColor} tick={{ fontSize: 10 }} interval="preserveStartEnd" />
                <YAxis stroke={axisColor} tick={{ fontSize: 10 }} allowDecimals={false} tickFormatter={(v: number) => formatBytes(v)} />
                <Tooltip contentStyle={tooltipStyle} formatter={(v: number) => [`${formatBytes(v)}/s`]} />
                <Legend wrapperStyle={{ fontSize: 11 }} />
                <Line type="monotone" dataKey="bytesIn" name="In" stroke="#f97316" strokeWidth={2} dot={false} isAnimationActive={false} />
                <Line type="monotone" dataKey="bytesOut" name="Out" stroke="#a78bfa" strokeWidth={2} dot={false} isAnimationActive={false} />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card className="border-flux-card-border bg-flux-card lg:col-span-2">
          <CardHeader className="pb-2">
            <CardTitle className="text-base text-flux-text">Active Connections</CardTitle>
            <p className="text-xs text-flux-text-muted">
              {scopeLabel} · connected clients over time
            </p>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={displayHistory} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
                <XAxis dataKey="time" stroke={axisColor} tick={{ fontSize: 10 }} interval="preserveStartEnd" />
                <YAxis stroke={axisColor} tick={{ fontSize: 10 }} allowDecimals={false} />
                <Tooltip contentStyle={tooltipStyle} />
                <Line type="monotone" dataKey="sessions" name="Connections" stroke="#3b82f6" strokeWidth={2} dot={false} isAnimationActive={false} />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

      </div>

      {nodes.length > 0 && (
        <Card className="border-flux-card-border bg-flux-card">
          <CardHeader className="pb-3">
            <CardTitle className="text-base text-flux-text">Cluster Nodes</CardTitle>
            <p className="text-xs text-flux-text-muted mt-0.5">{nodes.length} node{nodes.length !== 1 ? 's' : ''} · click a row to inspect</p>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow className="border-flux-card-border hover:bg-transparent">
                  <TableHead className="pl-6">Node</TableHead>
                  <TableHead>Address</TableHead>
                  <TableHead className="text-right">Sessions</TableHead>
                  <TableHead className="text-right">Subscriptions</TableHead>
                  <TableHead className="text-right">Msgs In</TableHead>
                  <TableHead className="text-right">Msgs Out</TableHead>
                  <TableHead className="text-right">Bytes In</TableHead>
                  <TableHead className="text-right pr-6">Uptime</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodes.map((node) => {
                  const isSelected = selectedNodeId === node.node_id;
                  return (
                    <TableRow
                      key={node.node_id}
                      className={`border-flux-card-border cursor-pointer transition-colors ${isSelected
                          ? 'bg-flux-blue/10 hover:bg-flux-blue/15'
                          : 'hover:bg-flux-hover'
                        }`}
                      onClick={() => setSelectedNodeId(isSelected ? null : node.node_id)}
                    >
                      <TableCell className="pl-6">
                        <div className="flex items-center gap-2">
                          <span className="inline-block w-2 h-2 rounded-full bg-flux-green shrink-0" />
                          <span className="text-flux-text font-medium text-sm">{node.node_id}</span>
                          {node.is_leader && (
                            <Badge variant="outline" className="text-xs bg-flux-blue/10 text-flux-blue border-flux-blue/20">
                              Leader
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-flux-text-muted font-mono text-xs">{node.addr}</TableCell>
                      <TableCell className="text-flux-text text-sm text-right">{formatCount(node.sessions)}</TableCell>
                      <TableCell className="text-flux-text text-sm text-right">{formatCount(node.subscriptions)}</TableCell>
                      <TableCell className="text-flux-text text-sm text-right">{formatCount(node.messages_received)}</TableCell>
                      <TableCell className="text-flux-text text-sm text-right">{formatCount(node.messages_sent)}</TableCell>
                      <TableCell className="text-flux-text text-sm text-right">{formatBytes(node.bytes_received)}</TableCell>
                      <TableCell className="text-flux-text-muted text-sm text-right pr-6">{formatUptime(node.uptime_seconds)}</TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

    </div>
  );
}

interface NodePillProps {
  label: string;
  active: boolean;
  isLeader?: boolean;
  onClick: () => void;
}

function NodePill({ label, active, isLeader, onClick }: NodePillProps) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-full border transition-colors ${active
          ? 'bg-flux-blue text-white border-flux-blue'
          : 'bg-flux-card text-flux-text-muted border-flux-card-border hover:bg-flux-hover hover:text-flux-text'
        }`}
    >
      {isLeader !== undefined && (
        <span className={`inline-block w-1.5 h-1.5 rounded-full ${active ? 'bg-white/70' : 'bg-flux-green'}`} />
      )}
      {label}
      {isLeader && (
        <span className={`text-[10px] ${active ? 'text-white/70' : 'text-flux-blue'}`}>★</span>
      )}
    </button>
  );
}
