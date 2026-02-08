import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { Button } from '../components/ui/Button';
import { api, SessionHistoryEntry, SessionHistorySummary } from '../lib/api';
import { getLocale, useI18n } from '../lib/i18n';
import { Clock3, Filter, ListChecks, Sigma } from 'lucide-react';

type PeriodFilter = 'today' | 'week' | 'month' | 'all';
type StatusFilter = 'completed' | 'abandoned' | 'all';

export const HistoryPage: React.FC = () => {
  const { language, t } = useI18n();
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  const [period, setPeriod] = useState<PeriodFilter>('month');
  const [status, setStatus] = useState<StatusFilter>('completed');
  const [tagsInput, setTagsInput] = useState('');
  const [minMinutesInput, setMinMinutesInput] = useState('');
  const [maxMinutesInput, setMaxMinutesInput] = useState('');
  const [loading, setLoading] = useState(true);
  const [entries, setEntries] = useState<SessionHistoryEntry[]>([]);
  const [summary, setSummary] = useState<SessionHistorySummary>({
    completed_count: 0,
    total_minutes: 0,
    average_minutes: 0,
  });

  const filters = useMemo(() => {
    const tags = tagsInput
      .split(',')
      .map((tag) => tag.trim())
      .filter(Boolean);
    const minMinutes = Number.parseInt(minMinutesInput, 10);
    const maxMinutes = Number.parseInt(maxMinutesInput, 10);
    return {
      period,
      status,
      tags,
      min_minutes: Number.isFinite(minMinutes) ? minMinutes : undefined,
      max_minutes: Number.isFinite(maxMinutes) ? maxMinutes : undefined,
      timezone,
      limit: 500,
    };
  }, [maxMinutesInput, minMinutesInput, period, status, tagsInput, timezone]);

  const fetchHistory = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.sessions.history(filters);
      setEntries(Array.isArray(data.sessions) ? data.sessions : []);
      const fallbackTotal = (data.sessions ?? []).reduce((sum, item) => sum + (item.recommended_minutes || 0), 0);
      const fallbackCount = data.sessions?.length ?? 0;
      setSummary(
        data.summary ?? {
          completed_count: fallbackCount,
          total_minutes: fallbackTotal,
          average_minutes: fallbackCount > 0 ? Math.round((fallbackTotal / fallbackCount) * 10) / 10 : 0,
        },
      );
    } catch (err) {
      console.error(err);
      setEntries([]);
      setSummary({ completed_count: 0, total_minutes: 0, average_minutes: 0 });
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    void fetchHistory();
  }, [fetchHistory]);

  return (
    <div className="space-y-8 font-sans">
      <header className="border-b border-border pb-4">
        <h1 className="text-2xl font-bold tracking-tight font-mono uppercase text-primary">{t('history.title')}</h1>
        <p className="mt-1 text-sm font-mono text-muted-foreground">{t('history.subtitle')}</p>
      </header>

      <Card className="border-border bg-surface shadow-none">
        <div className="mb-4 flex items-center gap-2">
          <Filter size={14} className="text-muted-foreground" />
          <span className="text-xs font-bold uppercase tracking-widest text-muted-foreground">{t('history.filters')}</span>
        </div>

        <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-6">
          <label className="space-y-1">
            <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('history.period')}</span>
            <select
              value={period}
              onChange={(event) => setPeriod(event.target.value as PeriodFilter)}
              className="h-10 w-full rounded-md border border-border bg-background px-3 text-xs font-mono text-primary"
            >
              <option value="today">{t('history.period.today')}</option>
              <option value="week">{t('history.period.week')}</option>
              <option value="month">{t('history.period.month')}</option>
              <option value="all">{t('history.period.all')}</option>
            </select>
          </label>

          <label className="space-y-1 lg:col-span-2">
            <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('history.tags')}</span>
            <Input
              value={tagsInput}
              onChange={(event) => setTagsInput(event.target.value)}
              placeholder={t('history.tagsPlaceholder')}
              className="text-xs font-mono"
            />
          </label>

          <label className="space-y-1">
            <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('history.minMinutes')}</span>
            <Input
              type="number"
              min={0}
              value={minMinutesInput}
              onChange={(event) => setMinMinutesInput(event.target.value)}
              placeholder="0"
              className="text-xs font-mono"
            />
          </label>

          <label className="space-y-1">
            <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('history.maxMinutes')}</span>
            <Input
              type="number"
              min={0}
              value={maxMinutesInput}
              onChange={(event) => setMaxMinutesInput(event.target.value)}
              placeholder="120"
              className="text-xs font-mono"
            />
          </label>

          <label className="space-y-1">
            <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('history.status')}</span>
            <select
              value={status}
              onChange={(event) => setStatus(event.target.value as StatusFilter)}
              className="h-10 w-full rounded-md border border-border bg-background px-3 text-xs font-mono text-primary"
            >
              <option value="completed">{t('history.status.completed')}</option>
              <option value="abandoned">{t('history.status.abandoned')}</option>
              <option value="all">{t('history.status.all')}</option>
            </select>
          </label>
        </div>

        <div className="mt-4 flex justify-end">
          <Button variant="outline" size="sm" onClick={() => void fetchHistory()}>
            {t('history.refresh')}
          </Button>
        </div>
      </Card>

      <div className="grid gap-4 md:grid-cols-3">
        <MetricCard icon={ListChecks} label={t('history.metric.closedCount')} value={summary.completed_count} />
        <MetricCard icon={Clock3} label={t('history.metric.totalMinutes')} value={summary.total_minutes} suffix="m" />
        <MetricCard icon={Sigma} label={t('history.metric.avgMinutes')} value={summary.average_minutes} suffix="m" />
      </div>

      <Card className="border-border bg-surface shadow-none">
        <h2 className="mb-4 text-xs font-bold uppercase tracking-widest text-muted-foreground">{t('history.listTitle')}</h2>
        {loading ? (
          <p className="text-sm text-muted-foreground">{t('history.loading')}</p>
        ) : entries.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t('history.empty')}</p>
        ) : (
          <div className="grid gap-2">
            {entries.map((entry) => (
              <div key={entry.session_id} className="rounded-lg border border-border bg-background/60 px-3 py-3">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate text-xs font-mono font-bold uppercase tracking-wide text-primary">{entry.topic}</p>
                    <p className="mt-1 truncate text-sm text-muted-foreground">{entry.desired_result}</p>
                  </div>
                  <span
                    className={`rounded px-2 py-0.5 text-[10px] font-mono uppercase ${
                      entry.status === 'completed'
                        ? 'bg-accent/15 text-accent'
                        : 'bg-red-500/10 text-red-400'
                    }`}
                  >
                    {entry.status === 'completed' ? t('history.status.completed') : t('history.status.abandoned')}
                  </span>
                </div>

                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px] font-mono uppercase text-muted-foreground">
                  <span>{entry.recommended_minutes}m</span>
                  <span>•</span>
                  <span>{new Date(entry.completed_at ?? entry.started_at).toLocaleDateString(getLocale(language), { month: 'short', day: 'numeric', year: 'numeric' })}</span>
                  {entry.tags?.map((tag) => (
                    <span key={`${entry.session_id}-${tag}`} className="rounded bg-surface px-1.5 py-0.5 text-[10px] text-muted-foreground">
                      #{tag}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
};

const MetricCard: React.FC<{ icon: any; label: string; value: number; suffix?: string }> = ({ icon: Icon, label, value, suffix }) => {
  return (
    <Card className="border-border bg-surface p-5 shadow-none">
      <div className="mb-2 flex h-8 w-8 items-center justify-center rounded bg-surfaceHighlight">
        <Icon size={15} className="text-primary" />
      </div>
      <p className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-mono font-bold text-primary">
        {value}
        {suffix ? <span className="ml-1 text-xs text-muted-foreground">{suffix}</span> : null}
      </p>
    </Card>
  );
};
