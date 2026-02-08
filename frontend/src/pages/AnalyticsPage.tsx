import React, { useEffect, useState } from 'react';
import { api, AnalyticsOverview, DailyActivity } from '../lib/api';
import { Card } from '../components/ui/Card';
import { Target, Zap, Activity, Award } from 'lucide-react';

export const AnalyticsPage: React.FC = () => {
  const [overview, setOverview] = useState<AnalyticsOverview | null>(null);
  const [activity, setActivity] = useState<DailyActivity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      api.analytics.overview(),
      api.analytics.activity(Intl.DateTimeFormat().resolvedOptions().timeZone)
    ]).then(([overviewData, activityData]) => {
      setOverview(overviewData.overview);
      setActivity(activityData.activity || []);
    }).catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="text-xs font-mono text-muted-foreground">LOADING ANALYTICS...</div>;

  return (
    <div className="space-y-8 font-sans">
      <header className="border-b border-border pb-4">
        <h1 className="text-2xl font-bold tracking-tight font-mono uppercase text-primary">Analytics</h1>
        <p className="text-sm text-muted-foreground font-mono mt-1">
          // PERFORMANCE METRICS
        </p>
      </header>

      {/* Overview Cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard 
          icon={Target} 
          label="Objectives Done" 
          value={overview?.completed_goals ?? 0} 
        />
        <StatCard 
          icon={Zap} 
          label="Focus Score" 
          value={overview?.calm_score ?? 0} 
          suffix="/ 100"
        />
        <StatCard 
          icon={Activity} 
          label="Stability" 
          value={overview?.focus_stability ?? 0} 
          suffix="%"
        />
        <StatCard 
          icon={Award} 
          label="Total Points" 
          value={overview?.total_nectar_earned ?? 0} 
          className="text-accent"
        />
      </div>

      {/* Heatmap */}
      <Card className="bg-surface border-border shadow-none">
        <h2 className="mb-4 text-xs font-bold uppercase tracking-widest text-muted-foreground">Activity Log (60 Days)</h2>
        <div className="flex flex-wrap gap-1">
          {Array.from({ length: 60 }).map((_, i) => {
            const date = new Date();
            date.setDate(date.getDate() - (59 - i));
            const dateStr = date.toISOString().split('T')[0];
            const dayActivity = activity.find(a => a.date === dateStr);
            const intensity = dayActivity 
              ? Math.min(dayActivity.total_minutes / 60, 1) 
              : 0;

            return (
              <div 
                key={i}
                title={`${dateStr}: ${dayActivity?.total_minutes || 0} mins`}
                className={`h-3 w-3 rounded-[2px] transition-all hover:ring-1 hover:ring-white ${
                  intensity === 0 ? 'bg-secondary' :
                  intensity < 0.3 ? 'bg-zinc-600' :
                  intensity < 0.7 ? 'bg-zinc-400' :
                  'bg-white'
                }`}
              />
            );
          })}
        </div>
        <div className="mt-4 flex items-center justify-end gap-2 text-[10px] font-mono uppercase text-muted-foreground">
          <span>Less</span>
          <div className="flex gap-1">
            <div className="h-2.5 w-2.5 rounded-[2px] bg-secondary" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-600" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-400" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-white" />
          </div>
          <span>More</span>
        </div>
      </Card>
    </div>
  );
};

const StatCard: React.FC<{ icon: any, label: string, value: number, suffix?: string, className?: string }> = ({ icon: Icon, label, value, suffix, className }) => (
  <Card className="flex flex-col items-start p-5 bg-surface border-border shadow-none transition-colors hover:bg-surfaceHighlight/50">
    <div className={`mb-3 flex h-8 w-8 items-center justify-center rounded bg-surfaceHighlight ${className}`}>
      <Icon size={16} className="text-primary" />
    </div>
    <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{label}</span>
    <div className="mt-1 flex items-baseline gap-1">
      <span className="text-2xl font-bold tracking-tight text-primary font-mono">{value}</span>
      {suffix && <span className="text-xs text-muted-foreground font-mono">{suffix}</span>}
    </div>
  </Card>
);