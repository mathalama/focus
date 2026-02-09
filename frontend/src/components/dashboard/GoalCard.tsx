import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useI18n } from '../../lib/i18n';
import { api } from '../../api';
import type { Goal } from '../../types';
import { Clock } from 'lucide-react';
import { motion } from 'framer-motion';

export const GoalCard: React.FC<{ goal: Goal }> = ({ goal }) => {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleStart = async () => {
    if (starting) return;
    setStarting(true);
    setError(null);
    try {
      const { session } = await api.sessions.start({
        goal_id: goal.id,
        recommended_minutes: goal.recommended_minutes,
        is_strict: false
      });
      navigate(`/session/${session.id}`);
    } catch (err) {
      console.error(err);
      setError(err instanceof Error ? err.message : t('goalCard.error'));
      setStarting(false);
    }
  };

  return (
    <motion.div
      whileHover={{ x: 4 }}
      transition={{ duration: 0.2 }}
    >
      <div 
        role="button"
        tabIndex={0}
        onClick={handleStart}
        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); handleStart(); } }}
        className="group relative flex cursor-pointer flex-col gap-2 rounded-lg border border-border bg-surface p-4 transition-colors hover:border-accent/50 hover:bg-surfaceHighlight"
      >
        <div className="flex items-start justify-between">
            <div className="flex flex-col gap-1">
                <h3 className="font-mono text-xs font-bold uppercase tracking-wide text-primary group-hover:text-accent transition-colors">
                    {goal.topic}
                </h3>
                <span className="text-[10px] text-muted-foreground font-mono">
                  {t('goalCard.id')}: {goal.id.substring(0, 8)}
                </span>
            </div>
            <div className="flex items-center gap-1 rounded bg-background px-1.5 py-0.5 text-[10px] font-mono font-medium text-muted-foreground border border-border">
              <Clock size={10} />
              <span>{goal.recommended_minutes}m</span>
            </div>
        </div>
          
        <p className="text-sm text-muted-foreground line-clamp-1 group-hover:text-primary transition-colors">
            {goal.desired_result}
        </p>

        {goal.tags && goal.tags.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-1">
              {goal.tags.map(tag => (
                <span key={tag} className="text-[9px] px-1 py-0 rounded bg-secondary text-muted-foreground font-mono uppercase">
                  #{tag}
                </span>
              ))}
            </div>
        )}

        {error && (
          <div className="mt-1 text-[11px] font-mono text-red-400">
            {error}
          </div>
        )}
      </div>
    </motion.div>
  );
};
