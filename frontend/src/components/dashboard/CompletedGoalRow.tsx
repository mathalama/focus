import React from 'react';
import { getLocale, useI18n } from '../../lib/i18n';
import type { Goal } from '../../types';
import { CheckCircle2, Clock } from 'lucide-react';

export const CompletedGoalRow: React.FC<{ goal: Goal }> = ({ goal }) => {
  const { language, t } = useI18n();
  const completedAt = goal.completed_at ? new Date(goal.completed_at) : null;
  const completedLabel = completedAt
    ? completedAt.toLocaleDateString(getLocale(language), { month: 'short', day: 'numeric' })
    : t('dashboard.completedFallback');

  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-surface/70 px-3 py-2">
      <div className="flex min-w-0 items-center gap-2">
        <CheckCircle2 size={14} className="text-accent" />
        <div className="min-w-0">
          <p className="truncate font-mono text-[11px] font-bold uppercase tracking-wide text-primary">{goal.topic}</p>
          <p className="truncate text-xs text-muted-foreground">{goal.desired_result}</p>
        </div>
      </div>

      <div className="ml-3 flex items-center gap-2 whitespace-nowrap text-[10px] font-mono uppercase text-muted-foreground">
        <Clock size={10} />
        <span>{goal.recommended_minutes}m</span>
        <span>•</span>
        <span>{completedLabel}</span>
      </div>
    </div>
  );
};
