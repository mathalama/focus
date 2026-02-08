import React, { useEffect, useMemo, useRef, useState } from 'react';
import { api, AnalyticsOverview, DailyActivity, DailyContribution } from '../lib/api';
import { Card } from '../components/ui/Card';
import { Target, Zap, Activity, Award } from 'lucide-react';
import { addDays, differenceInCalendarDays, eachDayOfInterval, format, startOfWeek, subDays } from 'date-fns';
import { getLocale, useI18n } from '../lib/i18n';

interface DayDetails {
  date: string;
  session_count: number;
  total_minutes: number;
  contributions: DailyContribution[];
}

interface HoverTooltip {
  visible: boolean;
  x: number;
  y: number;
  date: string;
  sessionCount: number;
  totalMinutes: number;
}

export const AnalyticsPage: React.FC = () => {
  const { t } = useI18n();
  const [overview, setOverview] = useState<AnalyticsOverview | null>(null);
  const [activity, setActivity] = useState<DailyActivity[]>([]);
  const [loading, setLoading] = useState(true);
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  useEffect(() => {
    Promise.all([
      api.analytics.overview(),
      api.analytics.activity(timezone)
    ]).then(([overviewData, activityData]) => {
      setOverview(overviewData.overview);
      setActivity(Array.isArray(activityData.activity) ? activityData.activity : []);
    }).catch(console.error)
      .finally(() => setLoading(false));
  }, [timezone]);

  if (loading) return <div className="text-xs font-mono text-muted-foreground">{t('analytics.loading')}</div>;

  // All date grid computations are done once per activity change (stable dep)
  return <AnalyticsContent overview={overview} activity={activity} />;
};

/**
 * Separated into its own component so that the heavy date-fns grid computation
 * only re-runs when `activity` actually changes (referential equality).
 */
const AnalyticsContent: React.FC<{
  overview: AnalyticsOverview | null;
  activity: DailyActivity[];
}> = ({ overview, activity }) => {
  const { language, t } = useI18n();
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [dayDetails, setDayDetails] = useState<DayDetails | null>(null);
  const [dayDetailsLoading, setDayDetailsLoading] = useState(false);
  const [tooltip, setTooltip] = useState<HoverTooltip>({
    visible: false,
    x: 0,
    y: 0,
    date: '',
    sessionCount: 0,
    totalMinutes: 0,
  });
  const hoverTimerRef = useRef<number | null>(null);
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  useEffect(() => () => {
    if (hoverTimerRef.current) {
      window.clearTimeout(hoverTimerRef.current);
    }
  }, []);

  const today = useMemo(() => new Date(), []);
  const rangeStart = useMemo(() => subDays(today, 364), [today]);

  const { weeks, activityByDate } = useMemo(() => {
    const gridStart = startOfWeek(rangeStart, { weekStartsOn: 0 });
    const totalDays = differenceInCalendarDays(today, gridStart) + 1;
    const allGridDays = eachDayOfInterval({ start: gridStart, end: addDays(gridStart, totalDays - 1) });

    const w: Date[][] = [];
    for (let i = 0; i < allGridDays.length; i += 7) {
      w.push(allGridDays.slice(i, i + 7));
    }

    const safeActivity = Array.isArray(activity) ? activity : [];
    const byDate = new Map(safeActivity.map((entry) => [entry.date, entry]));

    return { weeks: w, activityByDate: byDate };
  }, [today, rangeStart, activity]);

  const cellSize = 10;
  const cellGap = 3;
  const gridPixelWidth = weeks.length * cellSize + (weeks.length - 1) * cellGap;
  const weekDayColumnHeight = 7 * cellSize + 6 * cellGap;

  const monthMarkers = useMemo(() => {
    const markers: Array<{ weekIndex: number; label: string }> = [];
    let lastMonthKey = '';
    for (let weekIndex = 0; weekIndex < weeks.length; weekIndex += 1) {
      const monthStart = weeks[weekIndex].find((day) => day >= rangeStart && day <= today && day.getDate() === 1);
      if (monthStart) {
        const key = format(monthStart, 'yyyy-MM');
        if (key !== lastMonthKey) {
          markers.push({ weekIndex, label: format(monthStart, 'MMM') });
          lastMonthKey = key;
        }
      }
    }

    if (markers.length === 0 || markers[0].weekIndex !== 0 || format(rangeStart, 'MMM') !== markers[0].label) {
      markers.unshift({ weekIndex: 0, label: format(rangeStart, 'MMM') });
    }

    return markers;
  }, [rangeStart, today, weeks]);

  const intensityClass = (minutes: number, inRange: boolean) => {
    if (!inRange) return 'bg-transparent';
    if (minutes === 0) return 'bg-secondary';
    if (minutes < 25) return 'bg-zinc-700';
    if (minutes < 50) return 'bg-zinc-600';
    if (minutes < 90) return 'bg-zinc-500';
    return 'bg-zinc-400';
  };

  const formatLongDate = (dateKey: string) => {
    const parsed = new Date(`${dateKey}T00:00:00`);
    return parsed.toLocaleDateString(getLocale(language), {
      weekday: 'long',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const loadDayDetails = async (date: string) => {
    setSelectedDate(date);
    setDayDetailsLoading(true);
    try {
      const data = await api.analytics.day(date, timezone);
      const contributions = Array.isArray(data?.contributions) ? data.contributions : [];
      setDayDetails({
        date,
        session_count: Number.isFinite(data?.session_count) ? data.session_count : contributions.length,
        total_minutes: Number.isFinite(data?.total_minutes)
          ? data.total_minutes
          : contributions.reduce((acc: number, item: DailyContribution) => acc + (item.minutes || 0), 0),
        contributions,
      });
    } catch (err) {
      console.error(err);
      setDayDetails({ date, session_count: 0, total_minutes: 0, contributions: [] });
    } finally {
      setDayDetailsLoading(false);
    }
  };

  const clearHoverTimer = () => {
    if (hoverTimerRef.current) {
      window.clearTimeout(hoverTimerRef.current);
      hoverTimerRef.current = null;
    }
  };

  const onCellMouseEnter = (
    event: React.MouseEvent<HTMLButtonElement>,
    dateKey: string,
    inRange: boolean,
    sessionCount: number,
    totalMinutes: number,
  ) => {
    if (!inRange) return;
    clearHoverTimer();
    const x = event.clientX;
    const y = event.clientY;
    hoverTimerRef.current = window.setTimeout(() => {
      setTooltip({
        visible: true,
        x,
        y,
        date: dateKey,
        sessionCount,
        totalMinutes,
      });
    }, 450);
  };

  const onCellMouseMove = (event: React.MouseEvent<HTMLButtonElement>) => {
    setTooltip((prev) => (prev.visible ? { ...prev, x: event.clientX, y: event.clientY } : prev));
  };

  const onCellMouseLeave = () => {
    clearHoverTimer();
    setTooltip((prev) => ({ ...prev, visible: false }));
  };

  return (
    <div className="space-y-8 font-sans relative">
      <header className="border-b border-border pb-4">
        <h1 className="text-2xl font-bold tracking-tight font-mono uppercase text-primary">{t('analytics.title')}</h1>
        <p className="text-sm text-muted-foreground font-mono mt-1">
          {t('analytics.subtitle')}
        </p>
      </header>

      {/* Overview Cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard 
          icon={Target} 
          label={t('analytics.objectivesDone')} 
          value={overview?.completed_goals ?? 0} 
        />
        <StatCard 
          icon={Zap} 
          label={t('analytics.focusScore')} 
          value={overview?.calm_score ?? 0} 
          suffix="/ 100"
        />
        <StatCard 
          icon={Activity} 
          label={t('analytics.stability')} 
          value={overview?.focus_stability ?? 0} 
          suffix="%"
        />
        <StatCard 
          icon={Award} 
          label={t('analytics.totalPoints')} 
          value={overview?.total_nectar_earned ?? 0} 
          className="text-accent"
        />
      </div>

      {/* Heatmap */}
      <Card className="bg-surface border-border shadow-none">
        <h2 className="mb-4 text-xs font-bold uppercase tracking-widest text-muted-foreground">{t('analytics.activityLog')}</h2>
        <div className="overflow-x-auto pb-1">
          <div className="w-max">
            <div className="mb-2 ml-8 relative h-4" style={{ width: `${gridPixelWidth}px` }}>
              {monthMarkers.map((marker) => (
                <span
                  key={`${marker.label}-${marker.weekIndex}`}
                  className="absolute text-[10px] font-mono uppercase tracking-wider text-muted-foreground"
                  style={{ left: `${marker.weekIndex * (cellSize + cellGap)}px` }}
                >
                  {marker.label}
                </span>
              ))}
            </div>

            <div className="flex items-start gap-2">
              <div className="mt-[1px] flex w-6 flex-col justify-between text-[10px] font-mono uppercase tracking-wider text-muted-foreground" style={{ height: `${weekDayColumnHeight}px` }}>
                <span>Mon</span>
                <span>Wed</span>
                <span>Fri</span>
              </div>

              <div className="flex" style={{ gap: `${cellGap}px` }}>
                {weeks.map((week, weekIndex) => (
                  <div key={`week-${weekIndex}`} className="flex flex-col" style={{ gap: `${cellGap}px` }}>
                    {week.map((day) => {
                      const inRange = day >= rangeStart && day <= today;
                      const dateKey = format(day, 'yyyy-MM-dd');
                      const dayActivity = activityByDate.get(dateKey);
                      const sessionCount = inRange ? dayActivity?.session_count ?? 0 : 0;
                      const totalMinutes = inRange ? dayActivity?.total_minutes ?? 0 : 0;
                      const isSelected = selectedDate === dateKey;

                      return (
                        <button
                          type="button"
                          key={dateKey}
                          onClick={() => {
                            if (inRange) {
                              void loadDayDetails(dateKey);
                            }
                          }}
                          onMouseEnter={(event) => onCellMouseEnter(event, dateKey, inRange, sessionCount, totalMinutes)}
                          onMouseMove={onCellMouseMove}
                          onMouseLeave={onCellMouseLeave}
                          className={`appearance-none border-0 p-0 h-2.5 w-2.5 rounded-[2px] transition-all hover:ring-1 hover:ring-accent/70 ${intensityClass(totalMinutes, inRange)} ${isSelected ? 'ring-1 ring-accent' : ''}`}
                          aria-label={`${dateKey}: ${t('analytics.contributions', { count: sessionCount })}`}
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
          <span>{t('analytics.less')}</span>
          <div className="flex gap-1">
            <div className="h-2.5 w-2.5 rounded-[2px] bg-secondary" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-700" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-600" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-500" />
            <div className="h-2.5 w-2.5 rounded-[2px] bg-zinc-400" />
          </div>
          <span>{t('analytics.more')}</span>
        </div>
      </Card>

      <Card className="bg-surface border-border shadow-none">
        <h2 className="mb-4 text-xs font-bold uppercase tracking-widest text-muted-foreground">{t('analytics.dayHistory')}</h2>
        {!selectedDate ? (
          <p className="text-sm text-muted-foreground">{t('analytics.dayHistoryHint')}</p>
        ) : dayDetailsLoading ? (
          <p className="text-sm text-muted-foreground">{t('analytics.loadingDay')}</p>
        ) : (
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2 text-xs font-mono uppercase tracking-wider text-muted-foreground">
              <span>{formatLongDate(selectedDate)}</span>
              <span>•</span>
              <span>{t('analytics.contributions', { count: dayDetails?.session_count ?? 0 })}</span>
              <span>•</span>
              <span>{t('analytics.minutesShort', { minutes: dayDetails?.total_minutes ?? 0 })}</span>
            </div>

            {!dayDetails || dayDetails.contributions.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t('analytics.noDayData')}</p>
            ) : (
              <div className="grid gap-2">
                {dayDetails.contributions.map((item) => (
                  <div key={item.session_id} className="flex items-center justify-between rounded-md border border-border bg-background/60 px-3 py-2">
                    <div className="min-w-0">
                      <p className="truncate text-xs font-mono font-bold uppercase tracking-wide text-primary">{item.topic}</p>
                      <p className="text-[11px] text-muted-foreground">
                        {new Date(item.started_at).toLocaleTimeString(getLocale(language), { hour: '2-digit', minute: '2-digit' })}
                      </p>
                    </div>
                    <span className="ml-3 text-xs font-mono uppercase text-muted-foreground">{t('analytics.minutesShort', { minutes: item.minutes })}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </Card>

      {tooltip.visible && (
        <div
          className="pointer-events-none fixed z-50 rounded-md border border-border bg-surface px-3 py-2 text-xs font-mono text-primary shadow-lg"
          style={{ left: tooltip.x + 12, top: tooltip.y + 12 }}
        >
          <p>{t('analytics.contributions', { count: tooltip.sessionCount })}</p>
          <p className="text-muted-foreground">{formatLongDate(tooltip.date)}</p>
          <p className="text-muted-foreground">{t('analytics.minutesShort', { minutes: tooltip.totalMinutes })}</p>
        </div>
      )}
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
