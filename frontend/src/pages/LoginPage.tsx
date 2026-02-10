import React, { useMemo, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { api } from '../api';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Activity, ArrowRight } from 'lucide-react';
import { useI18n } from '../lib/i18n';

export const LoginPage: React.FC = () => {
  const { t } = useI18n();
  const location = useLocation();
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const { login } = useAuth();
  const navigate = useNavigate();
  const verifyState = useMemo(() => {
    const query = new URLSearchParams(location.search);
    const verified = query.get('verified');
    const reason = query.get('reason');
    if (verified === '1') {
      return { tone: 'success' as const, text: t('login.verifySuccess') };
    }
    if (verified === '0') {
      return { tone: 'error' as const, text: reason?.trim() ? reason.trim() : t('login.verifyFailed') };
    }
    return null;
  }, [location.search, t]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setNotice(null);
    try {
      if (mode === 'register') {
        const payload = await api.auth.register(email, name, password);
        setNotice(payload?.verification_email_sent ? t('login.registerSuccess') : t('login.registerCreatedButEmailFailed'));
        setMode('login');
        setPassword('');
        return;
      }

      const { user, token, refresh_token } = await api.auth.login(email, password);
      login(user, token, refresh_token);
      navigate('/');
    } catch (err) {
      console.error('Login failed', err);
      setError(err instanceof Error ? err.message : t('login.error'));
    } finally {
      setLoading(false);
    }
  };

  const handleResendVerification = async () => {
    setResending(true);
    setError(null);
    setNotice(null);
    try {
      await api.auth.resendVerification(email);
      setNotice(t('login.resendSuccess'));
    } catch (err) {
      setError(err instanceof Error ? err.message : t('login.resendError'));
    } finally {
      setResending(false);
    }
  };

  const unverifiedError = (error ?? '').toLowerCase().includes('not verified');

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4 font-sans text-primary">
      <Card className="w-full max-w-sm border-border bg-surface p-8 shadow-none">
        <div className="mb-8 flex flex-col items-start">
          <div className="mb-6 flex h-10 w-10 items-center justify-center rounded bg-accent text-accent-foreground">
            <Activity size={20} />
          </div>
          <h1 className="text-xl font-bold tracking-tight font-mono uppercase">MathalamaFocus</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {t('login.subtitle')}
          </p>
        </div>

        {verifyState ? (
          <div className={`mb-4 rounded-lg border px-4 py-2 text-xs font-mono ${verifyState.tone === 'success' ? 'border-emerald-500/30 bg-emerald-950/20 text-emerald-400' : 'border-red-500/30 bg-red-950/20 text-red-400'}`}>
            {verifyState.text}
          </div>
        ) : null}

        <div className="mb-5 grid grid-cols-2 gap-2 rounded-lg bg-background p-1">
          <button
            type="button"
            className={`h-9 rounded-md text-xs font-mono uppercase transition-colors ${mode === 'login' ? 'bg-surfaceHighlight text-primary' : 'text-muted-foreground hover:text-primary'}`}
            onClick={() => {
              setMode('login');
              setError(null);
              setNotice(null);
            }}
          >
            {t('login.mode.login')}
          </button>
          <button
            type="button"
            className={`h-9 rounded-md text-xs font-mono uppercase transition-colors ${mode === 'register' ? 'bg-surfaceHighlight text-primary' : 'text-muted-foreground hover:text-primary'}`}
            onClick={() => {
              setMode('register');
              setError(null);
              setNotice(null);
            }}
          >
            {t('login.mode.register')}
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div className="space-y-1.5">
            <label htmlFor="email" className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t('login.email')}
            </label>
            <Input
              id="email"
              type="email"
              placeholder={t('login.placeholder.email')}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="bg-background border-border text-primary placeholder:text-muted/20"
              required
              autoFocus
            />
          </div>

          {mode === 'register' ? (
            <div className="space-y-1.5">
              <label htmlFor="name" className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                {t('login.name')}
              </label>
              <Input
                id="name"
                type="text"
                placeholder={t('login.placeholder.name')}
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="bg-background border-border text-primary placeholder:text-muted/20"
                required
              />
            </div>
          ) : null}

          <div className="space-y-1.5">
            <label htmlFor="password" className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t('login.password')}
            </label>
            <Input
              id="password"
              type="password"
              placeholder={t('login.placeholder.password')}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="bg-background border-border text-primary placeholder:text-muted/20"
              required
            />
          </div>

          {error && (
            <div className="rounded-lg border border-red-500/30 bg-red-950/20 px-4 py-2 text-xs font-mono text-red-400 animate-in fade-in">
              {error}
            </div>
          )}

          {notice && (
            <div className="rounded-lg border border-emerald-500/30 bg-emerald-950/20 px-4 py-2 text-xs font-mono text-emerald-400 animate-in fade-in">
              {notice}
            </div>
          )}

          {mode === 'login' && unverifiedError ? (
            <Button
              type="button"
              variant="outline"
              className="w-full"
              onClick={handleResendVerification}
              disabled={resending || !email.trim()}
            >
              {resending ? t('login.resending') : t('login.resendVerification')}
            </Button>
          ) : null}

          <Button 
            type="submit" 
            className="w-full justify-between bg-accent text-accent-foreground hover:bg-accent/90 mt-2" 
            size="md" 
            disabled={loading}
          >
            <span>{loading ? t('login.authenticating') : mode === 'register' ? t('login.createAccount') : t('login.signIn')}</span>
            {!loading && <ArrowRight size={16} />}
          </Button>
        </form>
      </Card>
    </div>
  );
};
