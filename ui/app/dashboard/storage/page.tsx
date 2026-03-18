'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const storageMetrics = [
  { name: 'Message Queue',     size: '2.5 GB', itemCount: '125,432', usage: 65 },
  { name: 'Topic Index',       size: '1.2 GB', itemCount: '234',     usage: 42 },
  { name: 'Session Store',     size: '856 MB', itemCount: '892',     usage: 28 },
  { name: 'Retained Messages', size: '342 MB', itemCount: '5,234',   usage: 18 },
];

const StoragePage = () => {
  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-flux-text mb-1">Storage</h1>
        <p className="text-flux-text-muted">Monitor message storage and database metrics</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {storageMetrics.map((metric, i) => (
          <Card key={i} className="border-flux-card-border bg-flux-card">
            <CardContent className="p-5">
              <p className="text-flux-text-muted text-xs mb-1">{metric.name}</p>
              <p className="text-2xl font-bold text-flux-text mb-1">{metric.size}</p>
              <p className="text-xs text-flux-text-muted mb-3">{metric.itemCount} items</p>
              <div className="w-full bg-flux-bg rounded-full h-1.5">
                <div
                  className="bg-gradient-to-r from-flux-blue to-flux-teal h-1.5 rounded-full"
                  style={{ width: `${metric.usage}%` }}
                />
              </div>
              <p className="text-xs text-flux-text-muted mt-1.5">{metric.usage}% used</p>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card className="border-flux-card-border bg-flux-card">
        <CardHeader>
          <CardTitle className="text-base text-flux-text">Disk Space</CardTitle>
        </CardHeader>
        <CardContent>
          <div>
            <div className="flex justify-between mb-2">
              <span className="text-flux-text-muted text-sm">Total Used</span>
              <span className="text-flux-text font-semibold text-sm">4.9 GB / 10 GB</span>
            </div>
            <div className="w-full bg-flux-bg rounded-full h-2">
              <div className="bg-gradient-to-r from-flux-green to-flux-teal h-2 rounded-full" style={{ width: '49%' }} />
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="border-flux-card-border bg-flux-card">
          <CardHeader>
            <CardTitle className="text-base text-flux-text">Database Information</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {[
                { label: 'Type',        value: 'BadgerDB' },
                { label: 'Compression', value: 'Snappy' },
                { label: 'Data Dir',    value: '/tmp/fluxmq/data' },
                { label: 'Sync Writes', value: 'Disabled' },
              ].map(({ label, value }) => (
                <div key={label} className="flex justify-between">
                  <span className="text-flux-text-muted text-sm">{label}</span>
                  <span className="text-flux-text text-sm font-medium">{value}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-flux-card-border bg-flux-card">
          <CardHeader>
            <CardTitle className="text-base text-flux-text">Queue Stats</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {[
                { label: 'Queue Name',  value: 'mqtt' },
                { label: 'Topics',      value: '$queue/#' },
                { label: 'Max Depth',   value: '100,000' },
                { label: 'Message TTL', value: '7 days' },
              ].map(({ label, value }) => (
                <div key={label} className="flex justify-between">
                  <span className="text-flux-text-muted text-sm">{label}</span>
                  <span className="text-flux-text text-sm font-medium">{value}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default StoragePage;
