import React, { useEffect, useState } from 'react';
import { Shield, RefreshCw } from 'lucide-react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { api } from '../api';

type AdminEvent = {
  id: string;
  event_name: string;
  user_id?: string;
  created_at: string;
};

type EmailDelivery = {
  id: string;
  to_email: string;
  status: string;
  attempts: number;
  created_at: string;
};

export const AdminPage: React.FC = () => {
  const [events, setEvents] = useState<AdminEvent[]>([]);
  const [deliveries, setDeliveries] = useState<EmailDelivery[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const [eventsRes, deliveriesRes] = await Promise.all([
        api.admin.events(20),
        api.admin.emailDeliveries(20),
      ]);
      setEvents(Array.isArray(eventsRes?.events) ? eventsRes.events : []);
      setDeliveries(Array.isArray(deliveriesRes?.deliveries) ? deliveriesRes.deliveries : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load admin data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  return (
    <section className="space-y-4">
      <Card className="border-border bg-surface p-6">
        <div className="mb-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded bg-accent/20 p-2 text-accent">
              <Shield size={18} />
            </div>
            <h1 className="text-sm font-mono uppercase tracking-wider text-primary">Admin Console</h1>
          </div>
          <Button size="sm" variant="outline" onClick={() => void load()} disabled={loading}>
            <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
            Refresh
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">Outbox statuses and product events overview.</p>
        {error ? <p className="mt-2 text-xs text-red-400">{error}</p> : null}
      </Card>

      <Card className="border-border bg-surface p-4">
        <h2 className="mb-3 text-xs font-mono uppercase tracking-wider text-muted-foreground">Recent Email Deliveries</h2>
        <div className="space-y-2 text-xs">
          {deliveries.map((item) => (
            <div key={item.id} className="rounded border border-border/70 p-2">
              <div className="flex items-center justify-between">
                <span className="font-mono text-primary">{item.to_email}</span>
                <span className="text-muted-foreground">{item.status}</span>
              </div>
              <div className="text-[11px] text-muted-foreground">attempts: {item.attempts}</div>
            </div>
          ))}
          {!deliveries.length ? <p className="text-muted-foreground">No deliveries yet.</p> : null}
        </div>
      </Card>

      <Card className="border-border bg-surface p-4">
        <h2 className="mb-3 text-xs font-mono uppercase tracking-wider text-muted-foreground">Recent Product Events</h2>
        <div className="space-y-2 text-xs">
          {events.map((event) => (
            <div key={event.id} className="rounded border border-border/70 p-2">
              <div className="font-mono text-primary">{event.event_name}</div>
              <div className="text-[11px] text-muted-foreground">
                {event.user_id ? `user=${event.user_id}` : 'user=anonymous'}
              </div>
            </div>
          ))}
          {!events.length ? <p className="text-muted-foreground">No events yet.</p> : null}
        </div>
      </Card>
    </section>
  );
};
