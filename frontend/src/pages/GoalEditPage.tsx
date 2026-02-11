import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useI18n } from '../lib/i18n';
import { api } from '../api';
import type { Goal } from '../types';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { motion } from 'framer-motion';
import { ArrowLeft } from 'lucide-react';

export const GoalEditPage: React.FC = () => {
  const { goalId } = useParams<{ goalId: string }>();
  const navigate = useNavigate();
  const { t } = useI18n();
  const [goal, setGoal] = useState<Goal | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    topic: '',
    desired_result: '',
    recommended_minutes: 30,
    tags: [] as string[]
  });
  const [tagInput, setTagInput] = useState('');

  useEffect(() => {
    const loadGoal = async () => {
      if (!goalId) return;
      try {
        const data = await api.goals.get(goalId);
        setGoal(data);
        setFormData({
          topic: data.topic,
          desired_result: data.desired_result,
          recommended_minutes: data.recommended_minutes,
          tags: data.tags || []
        });
      } catch (err) {
        console.error(err);
        setError(err instanceof Error ? err.message : t('goalEdit.loadError'));
      } finally {
        setLoading(false);
      }
    };

    loadGoal();
  }, [goalId, t]);

  const handleAddTag = () => {
    const trimmedTag = tagInput.trim().toLowerCase();
    if (trimmedTag && !formData.tags.includes(trimmedTag)) {
      setFormData(prev => ({
        ...prev,
        tags: [...prev.tags, trimmedTag]
      }));
      setTagInput('');
    }
  };

  const handleRemoveTag = (tag: string) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.filter(t => t !== tag)
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (saving || !goalId) return;

    setSaving(true);
    setError(null);

    try {
      await api.goals.update(goalId, {
        topic: formData.topic,
        desired_result: formData.desired_result,
        recommended_minutes: formData.recommended_minutes,
        tags: formData.tags
      });
      navigate('/');
    } catch (err) {
      console.error(err);
      setError(err instanceof Error ? err.message : t('goalEdit.saveError'));
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-[400px] items-center justify-center text-xs font-mono text-muted-foreground">
        {t('goalEdit.loading')}
      </div>
    );
  }

  if (!goal) {
    return (
      <div className="space-y-4">
        <button
          onClick={() => navigate('/')}
          className="flex items-center gap-2 text-sm text-accent hover:text-accent/80 transition-colors"
        >
          <ArrowLeft size={16} />
          {t('goalEdit.back')}
        </button>
        <Card className="p-8 text-center text-sm text-muted-foreground">
          {t('goalEdit.notFound')}
        </Card>
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-2xl space-y-6"
    >
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate('/')}
          className="flex items-center gap-2 text-sm text-accent hover:text-accent/80 transition-colors"
        >
          <ArrowLeft size={16} />
          {t('goalEdit.back')}
        </button>
        <h1 className="text-2xl font-bold tracking-tight font-mono uppercase text-primary">
          {t('goalEdit.title')}
        </h1>
        <div className="w-[80px]" />
      </div>

      <Card className="p-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="p-3 rounded bg-red-500/10 border border-red-500/30 text-sm text-red-400 font-mono">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <label className="block text-xs font-mono font-bold uppercase tracking-wider text-muted-foreground">
              {t('goalEdit.topic')}
            </label>
            <input
              type="text"
              value={formData.topic}
              onChange={(e) => setFormData(prev => ({ ...prev, topic: e.target.value }))}
              placeholder={t('goalEdit.topicPlaceholder')}
              className="w-full rounded border border-border bg-background px-3 py-2 text-sm font-mono text-primary placeholder-muted-foreground focus:border-accent focus:outline-none"
              required
            />
          </div>

          <div className="space-y-2">
            <label className="block text-xs font-mono font-bold uppercase tracking-wider text-muted-foreground">
              {t('goalEdit.desiredResult')}
            </label>
            <textarea
              value={formData.desired_result}
              onChange={(e) => setFormData(prev => ({ ...prev, desired_result: e.target.value }))}
              placeholder={t('goalEdit.desiredResultPlaceholder')}
              rows={3}
              className="w-full rounded border border-border bg-background px-3 py-2 text-sm font-mono text-primary placeholder-muted-foreground focus:border-accent focus:outline-none resize-none"
              required
            />
          </div>

          <div className="space-y-2">
            <label className="block text-xs font-mono font-bold uppercase tracking-wider text-muted-foreground">
              {t('goalEdit.recommendedMinutes')}
            </label>
            <input
              type="number"
              value={formData.recommended_minutes}
              onChange={(e) => setFormData(prev => ({ ...prev, recommended_minutes: Math.max(1, parseInt(e.target.value) || 1) }))}
              min="1"
              max="480"
              className="w-full rounded border border-border bg-background px-3 py-2 text-sm font-mono text-primary focus:border-accent focus:outline-none"
              required
            />
          </div>

          <div className="space-y-2">
            <label className="block text-xs font-mono font-bold uppercase tracking-wider text-muted-foreground">
              {t('goalEdit.tags')}
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    handleAddTag();
                  }
                }}
                placeholder={t('goalEdit.tagsPlaceholder')}
                className="flex-1 rounded border border-border bg-background px-3 py-2 text-sm font-mono text-primary placeholder-muted-foreground focus:border-accent focus:outline-none"
              />
              <Button
                type="button"
                onClick={handleAddTag}
                variant="outline"
                size="sm"
              >
                {t('goalEdit.addTag')}
              </Button>
            </div>
            {formData.tags.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-2">
                {formData.tags.map((tag) => (
                  <div
                    key={tag}
                    className="flex items-center gap-2 px-2 py-1 rounded bg-secondary text-muted-foreground text-xs font-mono uppercase"
                  >
                    #{tag}
                    <button
                      type="button"
                      onClick={() => handleRemoveTag(tag)}
                      className="ml-1 hover:text-primary transition-colors"
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="flex gap-3 pt-4 border-t border-border">
            <Button
              type="submit"
              disabled={saving}
              className="flex-1"
              variant="primary"
            >
              {saving ? t('goalEdit.saving') : t('goalEdit.save')}
            </Button>
            <Button
              type="button"
              onClick={() => navigate('/')}
              variant="outline"
              className="flex-1"
            >
              {t('goalEdit.cancel')}
            </Button>
          </div>
        </form>
      </Card>
    </motion.div>
  );
};
