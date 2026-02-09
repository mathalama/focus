import React from 'react';
import { useAuth } from '../context/AuthContext';
import { AppLanguage, useLanguage } from '../context/LanguageContext';
import { Card } from '../components/ui/Card';
import { Globe, UserRound } from 'lucide-react';
import { useI18n } from '../lib/i18n';

const languageOptions: Array<{ value: AppLanguage; label: string }> = [
  { value: 'en', label: 'English' },
  { value: 'ru', label: 'Русский' },
  { value: 'kk', label: 'Қазақша' },
];

export const ProfilePage: React.FC = () => {
  const { user } = useAuth();
  const { language, setLanguage } = useLanguage();
  const { t } = useI18n();

  return (
    <div className="space-y-8 font-sans">
      <header className="border-b border-border pb-4">
        <h1 className="text-2xl font-bold tracking-tight font-mono uppercase text-primary">{t('profile.title')}</h1>
        <p className="mt-1 text-sm font-mono text-muted-foreground">{t('profile.subtitle')}</p>
      </header>

      <Card className="border-border bg-surface p-6 shadow-none">
        <div className="mb-6 flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded bg-surfaceHighlight text-accent">
            <UserRound size={18} />
          </div>
          <div>
            <p className="font-mono text-sm font-bold uppercase tracking-wide text-primary">{user?.name}</p>
            <p className="text-xs text-muted-foreground">{user?.email}</p>
          </div>
        </div>

        <div className="grid gap-2 text-xs font-mono text-muted-foreground sm:grid-cols-2">
          <span>{t('profile.currentPoints', { points: user?.nectar_balance ?? 0 })}</span>
          <span>{t('profile.totalEarned', { points: user?.total_nectar_earned ?? 0 })}</span>
        </div>
      </Card>

      <Card className="border-border bg-surface p-6 shadow-none">
        <div className="mb-4 flex items-center gap-2">
          <Globe size={16} className="text-accent" />
          <h2 className="font-mono text-sm font-bold uppercase tracking-wide text-primary">{t('profile.language')}</h2>
        </div>

        <label className="mb-2 block text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
          {t('profile.interfaceLanguage')}
        </label>
        <select
          value={language}
          onChange={(e) => setLanguage(e.target.value as AppLanguage)}
          className="h-11 w-full rounded-xl border-2 border-border bg-background px-4 text-sm text-primary outline-none transition-all focus:border-primary/20"
        >
          {languageOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>

        <p className="mt-3 text-xs text-muted-foreground">
          {t('profile.rollout')}
        </p>
      </Card>
    </div>
  );
};
