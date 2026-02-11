import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { useI18n } from '../lib/i18n';
import { Activity, ArrowLeft, CheckCircle } from 'lucide-react';

export const ForgotPasswordPage: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useI18n();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const trimmedEmail = email.trim();
      if (!trimmedEmail) {
        setError('Email is required');
        setLoading(false);
        return;
      }

      await api.auth.forgotPassword(trimmedEmail);
      setSent(true);
    } catch (err) {
      console.error('Forgot password error:', err);
      setError(err instanceof Error ? err.message : 'Failed to send reset email');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4 font-sans text-primary">
      <Card className="w-full max-w-sm border-border bg-surface p-8 shadow-none">
        <button
          onClick={() => navigate('/login')}
          className="mb-6 flex items-center gap-2 text-sm text-muted-foreground hover:text-primary transition-colors"
        >
          <ArrowLeft size={16} />
          {t('password.backLogin')}
        </button>

        <div className="mb-8 flex flex-col items-start">
          <div className="mb-6 flex h-10 w-10 items-center justify-center rounded bg-accent text-accent-foreground">
            <Activity size={20} />
          </div>
          <h1 className="text-xl font-bold tracking-tight font-mono uppercase">{t('password.resetTitle')}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {sent
              ? t('password.resetCheckEmail')
              : t('password.resetSubtitle')}
          </p>
        </div>

        {sent ? (
          <div className="flex flex-col items-center py-6">
            <CheckCircle size={48} className="mb-4 text-emerald-500" />
            <p className="text-center text-sm text-muted-foreground mb-4">
              We've sent password reset instructions to <strong>{email}</strong>
            </p>
            <p className="text-center text-xs text-muted-foreground mb-6">
              Check your spam folder if you don't see it within a few minutes.
            </p>
            <Button
              onClick={() => navigate('/login')}
              className="w-full"
            >
              {t('password.backLogin')}
            </Button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="rounded-lg border border-red-500/30 bg-red-950/20 px-4 py-2 text-xs font-mono text-red-400">
                {error}
              </div>
            )}

            <div>
              <label htmlFor="email" className="block text-xs font-mono uppercase text-muted-foreground mb-2">
                {t('login.email')}
              </label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder={t('login.placeholder.email')}
                required
                autoFocus
              />
            </div>

            <Button
              type="submit"
              disabled={loading}
              className="w-full"
            >
              {loading ? t('login.authenticating') : t('login.initializeSession')}
            </Button>

            <div className="text-center text-xs text-muted-foreground">
              {t('login.mode.login')}?{' '}
              <button
                type="button"
                onClick={() => navigate('/login')}
                className="font-mono underline hover:text-primary transition-colors"
              >
                {t('login.signIn')}
              </button>
            </div>
          </form>
        )}
      </Card>
    </div>
  );
};
