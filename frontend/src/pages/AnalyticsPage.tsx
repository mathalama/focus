import React, { useEffect, useState } from 'react';
import { api, AnalyticsOverview, DailyActivity } from '../lib/api';
import { Card } from '../components/ui/Card';
import { Target, Zap, Activity, Award } from 'lucide-react';
import { addDays, differenceInCalendarDays, eachDayOfInterval, format, startOfWeek, subDays } from 'date-fns';

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

  const today = new Date();
  const rangeStart = subDays(today, 364);
  const gridStart = startOfWeek(rangeStart, { weekStartsOn: 0 });
  const totalDays = differenceInCalendarDays(today, gridStart) + 1;
  const allGridDays = eachDayOfInterval({ start: gridStart, end: addDays(gridStart, totalDays - 1) });

  const weeks: Date[][] = [];
  for (let i = 0; i < allGridDays.length; i += 7) {
    weeks.push(allGridDays.slice(i, i + 7));
  }

  const activityByDate = new Map(activity.map((entry) => [entry.date, entry]));

  const monthLabels = new Map<number, string>();
  monthLabels.set(0, format(rangeStart, 'MMM'));
  for (let w = 0; w < weeks.length; w++) {
    const monthStartDay = weeks[w].find((day) => day >= rangeStart && day <= today && day.getDate() === 1);
    if (monthStartDay) {
      monthLabels.set(w, format(monthStartDay, 'MMM'));
    }
  }

  const intensityClass = (minutes: number, inRange: boolean) => {
    if (!inRange) return 'bg-transparent';
    if (minutes === 0) return 'bg-secondary';
    if (minutes < 25) return 'bg-zinc-700';
    if (minutes < 50) return 'bg-zinc-500';
    if (minutes < 90) return 'bg-zinc-300';
    return 'bg-white';
  };

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
        <h2 className="mb-4 text-xs font-bold uppercase tracking-widest text-muted-foreground">Activity Log (Last 12 Months)</h2>
        <div className="overflow-x-auto pb-2">
          <div className="min-w-[840px]">
            <div className="mb-2 grid gap-1 text-[10px] font-mono uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: `repeat(${weeks.length}, minmax(0, 1fr))` }}>
              {weeks.map((_, weekIndex) => (
                <div key={`month-${weekIndex}`} className="h-3 truncate">
                  {monthLabels.get(weekIndex) ?? ''}
                </div>
              ))}
            </div>

            <div className="flex items-start gap-2">
              <div className="mt-[1px] flex h-[84px] flex-col justify-between text-[10px] font-mono uppercase tracking-wider text-muted-foreground">
                <span>Mon</span>
                <span>Wed</span>
                <span>Fri</span>
              </div>

              <div className="flex gap-1">
                {weeks.map((week, weekIndex) => (
                  <div key={`week-${weekIndex}`} className="flex flex-col gap-1">
                    {week.map((day) => {
                      const inRange = day >= rangeStart && day <= today;
                      const dateKey = format(day, 'yyyy-MM-dd');
                      const dayActivity = activityByDate.get(dateKey);
                      const totalMinutes = inRange ? dayActivity?.total_minutes ?? 0 : 0;

                      return (
                        <div
                          key={dateKey}
                          title={`${dateKey}: ${totalMinutes} mins`}
                          className={`h-3 w-3 rounded-[2px] transition-all hover:ring-1 hover:ring-white/80 ${intensityClass(totalMinutes, inRange)}`}
                        />
                      );
                    })}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        <div className="mt-4 flex items-center justify-end gap-2 text-[10px] font-mono uppercase text-muted-foreground">
          <span>Less</span>
          <div className="flex gap-1">
            <div className="h-2.5 w-2.5 rounded-[2px] bg-secondary" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-700" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-500" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-300" />
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
