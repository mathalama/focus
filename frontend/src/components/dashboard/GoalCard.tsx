import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useI18n } from '../../lib/i18n';
import { api } from '../../api';
import type { Goal } from '../../types';
import { Clock, Edit2, Trash2, Play } from 'lucide-react';
import { motion } from 'framer-motion';

export const GoalCard: React.FC<{ goal: Goal; onUpdate?: () => void }> = ({ goal, onUpdate }) => {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [starting, setStarting] = useState(false);
  const [deleting, setDeleting] = useState(false);
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

  const handleDelete = async () => {
    if (deleting || !window.confirm(t('goalCard.confirmDelete') || 'Are you sure you want to delete this goal?')) return;
    setDeleting(true);
    setError(null);
    try {
      await api.goals.delete(goal.id);
      onUpdate?.();
    } catch (err) {
      console.error(err);
      setError(err instanceof Error ? err.message : t('goalCard.deleteError'));
      setDeleting(false);
    }
  };

  return (
    <motion.div
      whileHover={{ x: 4 }}
      transition={{ duration: 0.2 }}
    >
      <div 
        className="group relative flex cursor-pointer flex-col gap-2 rounded-lg border border-border bg-surface p-4 transition-colors hover:border-accent/50 hover:bg-surfaceHighlight"
      >
        <div className="flex items-start justify-between">
            <div className="flex flex-col gap-1 flex-1">
                <h3 className="font-mono text-xs font-bold uppercase tracking-wide text-primary group-hover:text-accent transition-colors">
                    {goal.topic}
                </h3>
                <span className="text-[10px] text-muted-foreground font-mono">
                  {t('goalCard.id')}: {goal.id.substring(0, 8)}
                </span>
            </div>
            <div className="flex items-center gap-2">
              <div className="flex items-center gap-1 rounded bg-background px-1.5 py-0.5 text-[10px] font-mono font-medium text-muted-foreground border border-border">
                <Clock size={10} />
                <span>{goal.recommended_minutes}m</span>
              </div>
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

        <div className="flex gap-2 mt-2 pt-2 border-t border-border/50">
          <button
            onClick={(e) => { e.stopPropagation(); handleStart(); }}
            disabled={starting}
            className="flex-1 flex items-center justify-center gap-1 text-xs font-mono px-2 py-1 rounded bg-accent/10 hover:bg-accent/20 text-accent disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <Play size={12} />
            {starting ? t('goal.starting') : t('goalCard.start') || 'Start'}
          </button>
          <button
            onClick={(e) => { e.stopPropagation(); navigate(`/goal/${goal.id}/edit`); }}
            className="flex items-center justify-center p-1.5 rounded bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 transition-colors"
            title={t('goal.editButton')}
          >
            <Edit2 size={14} />
          </button>
          <button
            onClick={(e) => { e.stopPropagation(); handleDelete(); }}
            disabled={deleting}
            className="flex items-center justify-center p-1.5 rounded bg-red-500/10 hover:bg-red-500/20 text-red-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            title={t('goal.deleteButton')}
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>
    </motion.div>
  );
};
