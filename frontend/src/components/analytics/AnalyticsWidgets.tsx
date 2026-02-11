import React from 'react';
import { useI18n } from '../../lib/i18n';
import { 
  Flame, 
  TrendingUp, 
  Rocket,
  Crown,
  Zap,
  Star,
  Target,
  ThumbsUp,
  Trophy
} from 'lucide-react';

interface StreakBadgeProps {
  days: number;
  label?: string;
}

/**
 * Display current streak (consecutive days with sessions)
 */
export const StreakBadge: React.FC<StreakBadgeProps> = ({ days, label }) => {
  const { t } = useI18n();
  const isHotStreak = days >= 7;
  const isBurningStreak = days >= 14;
  
  return (
    <div className={`inline-flex items-center gap-2 rounded-full px-4 py-2 border ${
      isBurningStreak 
        ? 'border-red-500/50 bg-red-950/30'
        : isHotStreak
        ? 'border-orange-500/50 bg-orange-950/30'
        : 'border-yellow-500/50 bg-yellow-950/30'
    }`}>
      <Flame 
        size={16} 
        className={
          isBurningStreak 
            ? 'text-red-400'
            : isHotStreak
            ? 'text-orange-400'
            : 'text-yellow-400'
        }
      />
      <span className="text-sm font-mono font-bold text-primary">
        {days} {label || t('streak.suffix')}
      </span>
    </div>
  );
};

interface ComparisonCardProps {
  current: number;
  previous: number;
  label: string;
  unit?: string;
}

/**
 * Show comparison with previous period
 */
export const ComparisonCard: React.FC<ComparisonCardProps> = ({ current, previous, label, unit = '' }) => {
  const diff = current - previous;
  const percentChange = previous > 0 ? ((diff / previous) * 100).toFixed(0) : 0;
  const isPositive = diff >= 0;

  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-surfaceHighlight p-3">
      <div>
        <p className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground">{label}</p>
        <p className="mt-1 text-sm font-mono font-bold text-primary">
          {current}{unit}
        </p>
      </div>
      <div className={`flex items-center gap-1 rounded px-2 py-1 ${
        isPositive ? 'bg-green-950/30 text-green-400' : 'bg-red-950/30 text-red-400'
      }`}>
        <TrendingUp size={14} className={isPositive ? '' : 'rotate-180'} />
        <span className="text-xs font-mono font-bold">
          {isPositive ? '+' : ''}{percentChange}%
        </span>
      </div>
    </div>
  );
};

interface AchievementBadgeProps {
  icon: React.ReactNode;
  title: string;
  description: string;
  unlocked: boolean;
}

/**
 * Display achievement/badge
 */
export const AchievementBadge: React.FC<AchievementBadgeProps> = ({ icon, title, description, unlocked }) => {
  return (
    <div className={`flex flex-col items-center rounded-lg border p-4 text-center transition-all ${
      unlocked
        ? 'border-accent/50 bg-accent/10'
        : 'border-border/50 bg-surface/50 opacity-50'
    }`}>
      <div className={`mb-2 flex items-center justify-center h-8 w-8 ${
        unlocked ? 'text-accent' : 'text-muted-foreground/50'
      }`}>
        {icon}
      </div>
      <p className="text-xs font-mono font-bold uppercase text-primary">{title}</p>
      <p className="mt-1 text-[10px] text-muted-foreground">{description}</p>
    </div>
  );
};

interface MotivationMessageProps {
  sessionsCompleted: number;
  streak: number;
  minutesTotal: number;
}

/**
 * Show motivating message based on user's data
 */
export const MotivationMessage: React.FC<MotivationMessageProps> = ({ sessionsCompleted, streak, minutesTotal }) => {
  const { t } = useI18n();
  let message = '';
  let Icon = Zap;
  let iconColor = 'text-yellow-400';

  if (sessionsCompleted === 0) {
    message = t('motivation.start');
    Icon = Rocket;
    iconColor = 'text-blue-400';
  } else if (streak >= 30) {
    message = t('motivation.legend', { days: streak });
    Icon = Crown;
    iconColor = 'text-yellow-400';
  } else if (streak >= 14) {
    message = t('motivation.hotStreak', { days: streak });
    Icon = Flame;
    iconColor = 'text-red-400';
  } else if (streak >= 7) {
    message = t('motivation.weekStreak', { days: streak });
    Icon = Star;
    iconColor = 'text-yellow-400';
  } else if (sessionsCompleted >= 100) {
    message = t('motivation.100sessions');
    Icon = Trophy;
    iconColor = 'text-yellow-400';
  } else if (minutesTotal >= 500) {
    message = t('motivation.500minutes', { minutes: minutesTotal });
    Icon = Zap;
    iconColor = 'text-yellow-400';
  } else if (streak >= 3) {
    message = t('motivation.3days', { days: streak });
    Icon = Target;
    iconColor = 'text-green-400';
  } else if (sessionsCompleted > 0) {
    message = t('motivation.started', { sessions: sessionsCompleted });
    Icon = ThumbsUp;
    iconColor = 'text-green-400';
  }

  return (
    <div className="rounded-lg border border-accent/50 bg-accent/10 p-4 flex items-center gap-3">
      <Icon size={24} className={`flex-shrink-0 ${iconColor}`} />
      <p className="text-sm text-primary font-mono">{message}</p>
    </div>
  );
};
