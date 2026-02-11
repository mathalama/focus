import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useI18n } from '../../lib/i18n';
import { api } from '../../api';
import type { SessionHistoryEntry } from '../../types';
import { X, Trash2, Eye } from 'lucide-react';
import { motion } from 'framer-motion';

export const SessionDetailModal: React.FC<{
  session: SessionHistoryEntry;
  isOpen: boolean;
  onClose: () => void;
  onUpdate: () => void;
}> = ({ session, isOpen, onClose, onUpdate }) => {
  const { language, t } = useI18n();
  const navigate = useNavigate();
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDelete = async () => {
    if (deleting || !window.confirm(t('history.modal.confirmDelete') || 'Delete this session?')) return;
    
    setDeleting(true);
    setError(null);
    
    try {
      await api.sessions.delete(session.session_id);
      setDeleting(false);
      onUpdate();
      onClose();
    } catch (err) {
      console.error(err);
      setError(err instanceof Error ? err.message : t('history.modal.deleteError'));
      setDeleting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
        className="fixed inset-0 z-40 bg-black/50 flex items-center justify-center p-4"
      >
        {/* Modal */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95, y: 20 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: 20 }}
          onClick={(e) => e.stopPropagation()}
          className="z-50 w-full max-w-lg max-h-[90vh] rounded-lg border border-border bg-surface shadow-2xl overflow-y-auto"
        >
        <div className="p-6 space-y-4">
        {/* Header */}
        <div className="mb-4 flex items-start justify-between pb-2 border-b border-border">
          <div className="flex-1">
            <h2 className="font-mono text-lg font-bold uppercase tracking-wide text-primary">
              {session.topic}
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">{session.desired_result}</p>
          </div>
          <button
            onClick={onClose}
            className="rounded p-1 hover:bg-secondary transition-colors flex-shrink-0 ml-4"
            title={t('modal.close')}
          >
            <X size={20} className="text-muted-foreground" />
          </button>
        </div>

        {/* Status Badge */}
        <div className="flex items-center gap-2">
          <span
            className={`rounded px-2 py-1 text-xs font-mono uppercase ${
              session.status === 'completed'
                ? 'bg-accent/15 text-accent'
                : 'bg-red-500/10 text-red-400'
            }`}
          >
            {session.status === 'completed' ? t('history.status.completed') : t('history.status.abandoned')}
          </span>
          <span className="text-xs font-mono text-muted-foreground">
            {new Date(session.completed_at ?? session.started_at).toLocaleDateString(language === 'ru' ? 'ru-RU' : language === 'kk' ? 'kk-KZ' : 'en-US', {
              month: 'short',
              day: 'numeric',
              year: 'numeric',
            })}
          </span>
        </div>

        {/* Details */}
        <div className="mb-6 space-y-3 rounded-lg bg-background p-4">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">{t('history.modal.duration')}</span>
            <span className="font-mono font-bold text-primary">{session.recommended_minutes}m</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">{t('history.modal.started')}</span>
            <span className="font-mono text-primary">
              {new Date(session.started_at).toLocaleTimeString(language === 'ru' ? 'ru-RU' : language === 'kk' ? 'kk-KZ' : 'en-US')}
            </span>
          </div>
          {session.tags && session.tags.length > 0 && (
            <div className="flex flex-col gap-1">
              <span className="text-sm text-muted-foreground">{t('history.modal.tags')}</span>
              <div className="flex flex-wrap gap-1">
                {session.tags.map((tag) => (
                  <span key={tag} className="rounded bg-secondary px-2 py-1 text-xs text-muted-foreground">
                    #{tag}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Error Message */}
        {error && (
          <div className="rounded bg-red-500/10 border border-red-500/30 p-3 text-sm text-red-400 font-mono">
            {error}
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-3">
          <button
            onClick={() => navigate(`/session/${session.session_id}/reflection`)}
            className="flex-1 flex items-center justify-center gap-2 rounded px-3 py-2 bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 text-sm font-mono uppercase transition-colors"
          >
            <Eye size={14} />
            {t('history.modal.viewReflections')}
          </button>
          <button
            onClick={handleDelete}
            disabled={deleting}
            className="flex-1 flex items-center justify-center gap-2 rounded px-3 py-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-sm font-mono uppercase disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <Trash2 size={14} />
            {deleting ? t('history.modal.deleting') : t('history.modal.delete')}
          </button>
        </div>

        {/* Close Button */}
        <button
          onClick={onClose}
          className="mt-4 w-full rounded px-3 py-2 bg-secondary/50 hover:bg-secondary text-sm font-mono uppercase text-muted-foreground transition-colors"
        >
          {t('history.modal.close')}
        </button>
        </div>
        </motion.div>
      </motion.div>
    </>
  );
};
