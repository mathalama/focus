import React, { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { api } from '../api';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { useI18n } from '../lib/i18n';
import { Activity, CheckCircle, AlertCircle } from 'lucide-react';

export const ResetPasswordPage: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useI18n();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');

  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [validatingToken, setValidatingToken] = useState(true);

  useEffect(() => {
    if (!token) {
      setError('Invalid reset link. Please request a new password reset.');
      setValidatingToken(false);
    } else {
      setValidatingToken(false);
    }
  }, [token]);

  const validatePassword = (password: string): string | null => {
    if (password.length < 8) {
      return 'Password must be at least 8 characters';
    }
    if (password.length > 72) {
      return 'Password must be at most 72 characters';
    }
    return null;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const passwordError = validatePassword(newPassword);
    if (passwordError) {
      setError(passwordError);
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (!token) {
      setError('Invalid reset link');
      return;
    }

    setLoading(true);
    try {
      await api.auth.resetPassword(token, newPassword);
      setSuccess(true);
    } catch (err) {
      console.error('Reset password error:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to reset password';
      if (errorMessage.includes('invalid') || errorMessage.includes('expired')) {
        setError('This reset link has expired or is invalid. Please request a new one.');
      } else {
        setError(errorMessage);
      }
    } finally {
      setLoading(false);
    }
  };

  if (validatingToken) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4 font-sans text-primary">
        <Card className="w-full max-w-sm border-border bg-surface p-8 shadow-none">
          <div className="animate-pulse space-y-4">
            <div className="h-4 bg-surfaceHighlight rounded"></div>
            <div className="h-10 bg-surfaceHighlight rounded"></div>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4 font-sans text-primary">
      <Card className="w-full max-w-sm border-border bg-surface p-8 shadow-none">
        <div className="mb-8 flex flex-col items-start">
          <div className="mb-6 flex h-10 w-10 items-center justify-center rounded bg-accent text-accent-foreground">
            <Activity size={20} />
          </div>
          <h1 className="text-xl font-bold tracking-tight font-mono uppercase">{t('password.setNewTitle')}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {success ? t('password.resetSuccess') : t('password.setNewSubtitle')}
          </p>
        </div>

        {success ? (
          <div className="flex flex-col items-center py-6">
            <CheckCircle size={48} className="mb-4 text-emerald-500" />
            <p className="text-center text-sm text-muted-foreground mb-6">
              Your password has been successfully reset.
            </p>
            <Button
              onClick={() => navigate('/login')}
              className="w-full"
            >
              Sign In with New Password
            </Button>
          </div>
        ) : !token ? (
          <div className="flex flex-col items-center py-6">
            <AlertCircle size={48} className="mb-4 text-red-500" />
            <p className="text-center text-sm text-muted-foreground mb-6">
              {error}
            </p>
            <Button
              onClick={() => navigate('/login')}
              variant="outline"
              className="w-full mb-2"
            >
              Back to Login
            </Button>
            <Button
              onClick={() => navigate('/forgot-password')}
              className="w-full"
            >
              Request New Reset Link
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
              <label htmlFor="newPassword" className="block text-xs font-mono uppercase text-muted-foreground mb-2">
                {t('login.password')}
              </label>
              <Input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder={t('password.enterNewPassword')}
                required
                autoFocus
              />
              <p className="mt-1 text-xs text-muted-foreground">
                8-72 characters
              </p>
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-xs font-mono uppercase text-muted-foreground mb-2">
                {t('login.password')}
              </label>
              <Input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder={t('password.confirmNewPassword')}
                required
              />
            </div>

            <Button
              type="submit"
              disabled={loading || !newPassword || !confirmPassword}
              className="w-full"
            >
              {loading ? t('login.authenticating') : t('password.resetTitle')}
            </Button>

            <div className="text-center text-xs text-muted-foreground">
              <button
                type="button"
                onClick={() => navigate('/login')}
                className="font-mono underline hover:text-primary transition-colors"
              >
                {t('password.backLogin')}
              </button>
            </div>
          </form>
        )}
      </Card>
    </div>
  );
};
