'use client';

import { Search } from 'lucide-react';
import { MOCK_CLIENTS } from '@/lib/api';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

const ConnectionsPage = () => {
  const connections = MOCK_CLIENTS.map((c) => ({
    remote: c.remote ?? '—',
    protocol: c.protocol ?? 'MQTT',
    clientId: c.client_id,
    connected: c.connected,
    connectedAt: c.connected_at,
  }));

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-flux-text mb-1">Connections</h1>
        <p className="text-flux-text-muted">Monitor all broker connections</p>
      </div>

      <Card className="border-flux-card-border bg-flux-card">
        <CardContent className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div className="relative flex-1 max-w-xs">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-flux-blue dark:text-flux-orange" size={18} />
              <Input
                type="text"
                placeholder="Search connections..."
                className="pl-10 bg-flux-bg border-flux-card-border text-flux-text placeholder:text-flux-text-muted focus-visible:ring-flux-blue"
              />
            </div>
          </div>

          <Table>
            <TableHeader>
              <TableRow className="border-flux-card-border hover:bg-transparent">
                <TableHead>Remote</TableHead>
                <TableHead>Protocol</TableHead>
                <TableHead>Client ID</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Connected At</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {connections.map((conn, i) => (
                <TableRow key={i} className="border-flux-card-border hover:bg-flux-hover">
                  <TableCell className="text-flux-text font-mono text-sm">{conn.remote}</TableCell>
                  <TableCell>
                    <Badge
                      variant="outline"
                      className="bg-flux-blue/10 text-flux-blue border-flux-blue/20"
                    >
                      {conn.protocol}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-flux-text text-sm">{conn.clientId}</TableCell>
                  <TableCell>
                    <Badge
                      variant="outline"
                      className={`inline-flex items-center gap-1.5 ${
                        conn.connected
                          ? 'bg-flux-green/10 text-flux-green border-flux-green/20'
                          : 'bg-flux-red/10 text-flux-red border-flux-red/20'
                      }`}
                    >
                      <span className={`w-1.5 h-1.5 rounded-full ${conn.connected ? 'bg-flux-green' : 'bg-flux-red'}`} />
                      {conn.connected ? 'Connected' : 'Disconnected'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-flux-text-muted text-sm">
                    {new Date(conn.connectedAt).toLocaleString()}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
};

export default ConnectionsPage;
