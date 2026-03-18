'use client';

import { Search } from 'lucide-react';
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

const topics = [
  { name: 'sensors/temperature', subscribers: 5, messagesPerMin: 120, retained: true, created: '2024-01-15' },
  { name: 'sensors/humidity',    subscribers: 3, messagesPerMin: 85,  retained: true, created: '2024-01-15' },
  { name: 'devices/online',      subscribers: 12, messagesPerMin: 240, retained: false, created: '2024-01-10' },
  { name: 'alerts/critical',     subscribers: 8, messagesPerMin: 45,  retained: false, created: '2024-01-20' },
  { name: 'system/status',       subscribers: 2, messagesPerMin: 30,  retained: true, created: '2024-01-01' },
];

const TopicsPage = () => {
  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-flux-text mb-1">Topics</h1>
        <p className="text-flux-text-muted">Browse and manage MQTT topics</p>
      </div>

      <Card className="border-flux-card-border bg-flux-card">
        <CardContent className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div className="relative flex-1 max-w-xs">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-flux-blue dark:text-flux-orange" size={18} />
              <Input
                type="text"
                placeholder="Search topics..."
                className="pl-10 bg-flux-bg border-flux-card-border text-flux-text placeholder:text-flux-text-muted focus-visible:ring-flux-blue"
              />
            </div>
          </div>

          <Table>
            <TableHeader>
              <TableRow className="border-flux-card-border hover:bg-transparent">
                <TableHead>Topic Name</TableHead>
                <TableHead>Subscribers</TableHead>
                <TableHead>Msg/Min</TableHead>
                <TableHead>Retained</TableHead>
                <TableHead>Created</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {topics.map((topic, i) => (
                <TableRow key={i} className="border-flux-card-border hover:bg-flux-hover">
                  <TableCell className="text-flux-text font-medium text-sm font-mono">{topic.name}</TableCell>
                  <TableCell className="text-flux-text text-sm">{topic.subscribers}</TableCell>
                  <TableCell className="text-flux-text text-sm">{topic.messagesPerMin}</TableCell>
                  <TableCell>
                    <Badge
                      variant="outline"
                      className={
                        topic.retained
                          ? 'bg-flux-green/10 text-flux-green border-flux-green/20'
                          : 'bg-flux-card-border text-flux-text-muted border-flux-card-border'
                      }
                    >
                      {topic.retained ? 'Yes' : 'No'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-flux-text-muted text-sm">{topic.created}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
};

export default TopicsPage;
