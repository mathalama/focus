import React, { useCallback, useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { useLanguage } from '../context/LanguageContext';
import { api } from '../api';
import type { Goal } from '../types';
import { DailyQuote, getDailyQuote, msUntilNextLocalDay } from '../lib/dailyQuotes';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { useNavigate } from 'react-router-dom';
import { Quote } from 'lucide-react';
import { motion } from 'framer-motion';
import { useI18n } from '../lib/i18n';
import { GoalForm } from '../components/dashboard/GoalForm';
import { GoalCard } from '../components/dashboard/GoalCard';
import { CompletedGoalRow } from '../components/dashboard/CompletedGoalRow';

export const DashboardPage: React.FC = () => {
  const { user } = useAuth();
  const { language } = useLanguage();
  const { t } = useI18n();
  const navigate = useNavigate();
  const [goals, setGoals] = useState<Goal[]>([]);
  const [goalHistory, setGoalHistory] = useState<Goal[]>([]);
  const [dailyQuote, setDailyQuote] = useState<DailyQuote>(() => getDailyQuote(language, new Date()));
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const [goalsData, historyData] = await Promise.all([
        api.goals.list(),
        api.goals.history()
      ]);
      setGoals(goalsData.goals || []);
      setGoalHistory(historyData.goals || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let isMounted = true;

    const restoreOrLoad = async () => {
      try {
        const active = await api.sessions.active();
        if (isMounted && active.session?.id) {
          navigate(`/session/${active.session.id}`, { replace: true });
          return;
        }
      } catch (err) {
        console.error(err);
      }
      if (isMounted) {
        await fetchData();
      }
    };

    void restoreOrLoad();
    return () => {
      isMounted = false;
    };
  }, [fetchData, navigate]);

  useEffect(() => {
    let timer: number | undefined;

    const scheduleQuoteRefresh = () => {
      const now = new Date();
      setDailyQuote(getDailyQuote(language, now));
      timer = window.setTimeout(scheduleQuoteRefresh, msUntilNextLocalDay(now));
    };

    scheduleQuoteRefresh();

    return () => {
      if (timer) {
        window.clearTimeout(timer);
      }
    };
  }, [language]);

  return (
    <div className="space-y-8 font-sans">
      <header className="flex items-center justify-between border-b border-border pb-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight font-mono uppercase text-primary">{t('dashboard.title')}</h1>
          <p className="text-sm text-muted-foreground font-mono">
            {t('dashboard.statusLine', { name: user?.name.toUpperCase() ?? '-' })}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={fetchData} className="hidden sm:flex">
          {t('dashboard.refresh')}
        </Button>
      </header>

      <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }}>
        <Card className="bg-surfaceHighlight/20 border-border p-4">
          <div className="flex items-start gap-3">
            <div className="rounded bg-accent/10 p-2 text-accent">
              <Quote size={16} />
            </div>
            <div>
              <h3 className="mb-1 text-xs font-bold uppercase tracking-widest text-muted-foreground">{t('dashboard.dailyQuote')}</h3>
              <p className="text-sm leading-relaxed text-primary">{dailyQuote.text}</p>
              <p className="mt-2 text-[11px] font-mono uppercase tracking-wider text-muted-foreground">{dailyQuote.author}</p>
            </div>
          </div>
        </Card>
      </motion.div>

      <div className="grid gap-8 lg:grid-cols-[1fr,1.5fr]">
        <div className="space-y-6">
           <GoalForm />
        </div>

        <div className="space-y-8">
          <section className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-medium tracking-wider text-muted-foreground uppercase">{t('dashboard.recentObjectives')}</h2>
              <span className="text-xs font-mono text-muted-foreground">{t('dashboard.activeCount', { count: goals.length })}</span>
            </div>

            {loading ? (
              <div className="py-12 text-center text-xs font-mono text-muted-foreground animate-pulse">
                {t('dashboard.loadingObjectives')}
              </div>
            ) : goals.length === 0 ? (
              <div className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
                {t('dashboard.noActive')}
              </div>
            ) : (
              <div className="grid gap-3">
                {goals.map((goal) => (
                  <GoalCard key={goal.id} goal={goal} onUpdate={fetchData} />
                ))}
              </div>
            )}
          </section>

          <section className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-medium tracking-wider text-muted-foreground uppercase">{t('dashboard.completedHistory')}</h2>
              <span className="text-xs font-mono text-muted-foreground">{t('dashboard.doneCount', { count: goalHistory.length })}</span>
            </div>

            {loading ? (
              <div className="py-8 text-center text-xs font-mono text-muted-foreground animate-pulse">
                {t('dashboard.loadingHistory')}
              </div>
            ) : goalHistory.length === 0 ? (
              <div className="rounded-xl border border-dashed border-border p-6 text-center text-sm text-muted-foreground">
                {t('dashboard.noCompleted')}
              </div>
            ) : (
              <div className="grid gap-2">
                {goalHistory.map((goal) => (
                  <CompletedGoalRow key={goal.id} goal={goal} />
                ))}
              </div>
            )}
          </section>
        </div>
      </div>
    </div>
  );
};
