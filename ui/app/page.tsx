import { FluxLogo } from '@/components/flux-logo';
import { Link } from 'lucide-react';

export default function Home() {
  return (
    <main className="flex items-center justify-center min-h-screen bg-flux-bg">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-4">
          <FluxLogo /> Dashboard
        </h1>
        <Link href="/dashboard" className="inline-flex items-center gap-2 text-flux-blue hover:text-flux-blue-hover font-medium">
          View Dashboard
        </Link>
        <p className="text-xl text-flux-text-muted mb-8">Click here to view dashboard</p>
        <div className="animate-pulse">
          <div className="h-2 w-2 rounded-full mx-auto" style={{ background: 'var(--flux-blue)' }} />
        </div>
      </div>
    </main>
  );
}
