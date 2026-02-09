import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../api';
import { useAuth } from '../context/AuthContext';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { motion } from 'framer-motion';
import { useI18n } from '../lib/i18n';

export const ReflectionPage: React.FC = () => {
  const { t } = useI18n();
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { refreshUser } = useAuth();
  const [learned, setLearned] = useState('');
  const [hard, setHard] = useState('');
  const [next, setNext] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!sessionId) return;
    setLoading(true);
    setError(null);
    try {
      await api.sessions.reflection(sessionId, {
        what_learned: learned,
        what_was_hard: hard,
        next_action: next
      });
      await refreshUser();
      navigate('/hive');
    } catch (err) {
      console.error(err);
      setError(err instanceof Error ? err.message : t('reflection.error'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-[80vh] items-center justify-center">
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        className="w-full max-w-lg"
      >
        <Card className="p-8">
          <header className="mb-6 text-center">
            <h1 className="text-2xl font-bold tracking-tight text-primary">{t('reflection.title')}</h1>
            <p className="mt-2 text-muted-foreground">
              {t('reflection.subtitle')}
            </p>
          </header>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label className="text-sm font-medium">{t('reflection.learned')}</label>
              <textarea
                className="flex min-h-[80px] w-full rounded-xl border-2 border-transparent bg-secondary/50 px-4 py-3 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:border-primary/20 focus-visible:bg-surface transition-all resize-none"
                placeholder={t('reflection.learnedPlaceholder')}
                value={learned}
                onChange={e => setLearned(e.target.value)}
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('reflection.hard')}</label>
              <textarea
                className="flex min-h-[80px] w-full rounded-xl border-2 border-transparent bg-secondary/50 px-4 py-3 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:border-primary/20 focus-visible:bg-surface transition-all resize-none"
                placeholder={t('reflection.hardPlaceholder')}
                value={hard}
                onChange={e => setHard(e.target.value)}
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('reflection.next')}</label>
              <Input
                placeholder={t('reflection.nextPlaceholder')}
                value={next}
                onChange={e => setNext(e.target.value)}
                required
              />
            </div>

            <Button type="submit" className="w-full" size="lg" disabled={loading}>
              {loading ? t('reflection.saving') : t('reflection.collect')}
            </Button>

            {error && (
              <div className="rounded-lg border border-red-500/30 bg-red-950/20 px-4 py-2 text-xs font-mono text-red-400 animate-in fade-in">
                {error}
              </div>
            )}
          </form>
        </Card>
      </motion.div>
    </div>
  );
};
