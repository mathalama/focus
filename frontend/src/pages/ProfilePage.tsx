import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { AppLanguage, useLanguage } from '../context/LanguageContext';
import { Card } from '../components/ui/Card';
import { Bell, Copy, ExternalLink, Globe, LoaderCircle, UserRound } from 'lucide-react';
import { useI18n } from '../lib/i18n';
import { api, TelegramIdentity } from '../lib/api';
import { Button } from '../components/ui/Button';

const languageOptions: Array<{ value: AppLanguage; label: string }> = [
  { value: 'en', label: 'English' },
  { value: 'ru', label: 'Русский' },
  { value: 'kk', label: 'Қазақша' },
];

export const ProfilePage: React.FC = () => {
  const { user } = useAuth();
  const { language, setLanguage } = useLanguage();
  const { t, locale } = useI18n();
  const [telegramIdentity, setTelegramIdentity] = useState<TelegramIdentity | null>(null);
  const [telegramCode, setTelegramCode] = useState<{ code: string; expiresAt: string } | null>(null);
  const [telegramNotice, setTelegramNotice] = useState<{ tone: 'success' | 'error'; text: string } | null>(null);
  const [isTelegramStatusLoading, setIsTelegramStatusLoading] = useState(true);
  const [isGeneratingCode, setIsGeneratingCode] = useState(false);
  const [isCopyingCode, setIsCopyingCode] = useState(false);
  const [isUnlinkingTelegram, setIsUnlinkingTelegram] = useState(false);
  const telegramBotUrl = (import.meta.env.VITE_TELEGRAM_BOT_URL ?? '').trim();
  const telegramBotHandleMatch = telegramBotUrl.match(/(?:t\.me\/|telegram\.me\/)([A-Za-z0-9_]+)/i);
  const telegramBotLabel = telegramBotHandleMatch ? `@${telegramBotHandleMatch[1]}` : telegramBotUrl;

  const telegramExpiresLabel = useMemo(() => {
    if (!telegramCode?.expiresAt) return '';

    const parsedDate = new Date(telegramCode.expiresAt);
    if (Number.isNaN(parsedDate.getTime())) {
      return telegramCode.expiresAt;
    }

    return parsedDate.toLocaleString(locale, {
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
  }, [telegramCode?.expiresAt, locale]);

  const telegramIdentityLabel = useMemo(() => {
    if (!telegramIdentity) return '';

    const username = telegramIdentity.telegram_username?.trim();
    if (username) {
      return `@${username}`;
    }

    const fullName = [telegramIdentity.telegram_first_name, telegramIdentity.telegram_last_name]
      .map((part) => part.trim())
      .filter(Boolean)
      .join(' ');
    if (fullName) {
      return fullName;
    }

    return String(telegramIdentity.telegram_user_id);
  }, [telegramIdentity]);

  const telegramLinkedAtLabel = useMemo(() => {
    if (!telegramIdentity?.linked_at) return '';
    const parsedDate = new Date(telegramIdentity.linked_at);
    if (Number.isNaN(parsedDate.getTime())) {
      return telegramIdentity.linked_at;
    }
    return parsedDate.toLocaleString(locale, {
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
  }, [telegramIdentity?.linked_at, locale]);

  const loadTelegramIdentity = useCallback(async () => {
    setIsTelegramStatusLoading(true);
    try {
      const data = await api.auth.getTelegramIdentity();
      setTelegramIdentity(data.identity ?? null);
    } catch {
      setTelegramNotice({
        tone: 'error',
        text: t('profile.telegram.notice.statusFailed'),
      });
    } finally {
      setIsTelegramStatusLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void loadTelegramIdentity();
  }, [loadTelegramIdentity]);

  const handleGenerateTelegramCode = async () => {
    setIsGeneratingCode(true);
    setTelegramNotice(null);
    try {
      const data = await api.auth.createTelegramLinkCode();
      setTelegramCode({ code: data.code, expiresAt: data.expires_at });
      setTelegramNotice({
        tone: 'success',
        text: telegramIdentity ? t('profile.telegram.notice.relinkCodeGenerated') : t('profile.telegram.notice.generated'),
      });
    } catch (error) {
      setTelegramNotice({
        tone: 'error',
        text: error instanceof Error && error.message.trim()
          ? error.message
          : t('profile.telegram.notice.generateFailed'),
      });
    } finally {
      setIsGeneratingCode(false);
    }
  };

  const handleUnlinkTelegram = async () => {
    setIsUnlinkingTelegram(true);
    setTelegramNotice(null);
    try {
      await api.auth.unlinkTelegram();
      setTelegramIdentity(null);
      setTelegramCode(null);
      setTelegramNotice({
        tone: 'success',
        text: t('profile.telegram.notice.unlinked'),
      });
    } catch (error) {
      setTelegramNotice({
        tone: 'error',
        text: error instanceof Error && error.message.trim()
          ? error.message
          : t('profile.telegram.notice.unlinkFailed'),
      });
    } finally {
      setIsUnlinkingTelegram(false);
    }
  };

  const handleCopyTelegramCommand = async () => {
    if (!telegramCode?.code) return;

    if (!navigator.clipboard || !window.isSecureContext) {
      setTelegramNotice({
        tone: 'error',
        text: t('profile.telegram.notice.copyFailed'),
      });
      return;
    }

    setIsCopyingCode(true);
    try {
      await navigator.clipboard.writeText(`/link ${telegramCode.code}`);
      setTelegramNotice({
        tone: 'success',
        text: t('profile.telegram.notice.copied'),
      });
    } catch {
      setTelegramNotice({
        tone: 'error',
        text: t('profile.telegram.notice.copyFailed'),
      });
    } finally {
      setIsCopyingCode(false);
    }
  };

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
        <div className="mb-4 flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#229ED9]/15 text-[#229ED9]">
            <svg viewBox="0 0 24 24" className="h-5 w-5 fill-current" aria-hidden="true">
              <path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.587 7.896-1.97 9.289c-.149.658-.538.82-1.09.512l-3.012-2.222-1.453 1.397c-.161.161-.296.296-.605.296l.216-3.066 5.582-5.045c.243-.216-.054-.337-.378-.121l-6.902 4.347-2.972-.929c-.646-.203-.658-.646.135-.956l11.617-4.479c.54-.203 1.01.121.832.977z" />
            </svg>
          </div>
          <div>
            <h2 className="font-mono text-sm font-bold uppercase tracking-wide text-primary">{t('profile.telegram.title')}</h2>
            <p className="text-xs text-muted-foreground">{t('profile.telegram.description')}</p>
          </div>
        </div>

        <div className="mb-5 flex items-start gap-2 rounded-xl border border-border/70 bg-background/60 p-3 text-xs text-muted-foreground">
          <Bell size={14} className="mt-[1px] text-accent" />
          <p>{t('profile.telegram.notificationsHint')}</p>
        </div>

        {telegramBotUrl ? (
          <a
            href={telegramBotUrl}
            target="_blank"
            rel="noreferrer"
            className="mb-5 flex items-center justify-between rounded-xl border border-[#229ED9]/30 bg-[#229ED9]/10 px-4 py-3 transition-colors hover:bg-[#229ED9]/15"
          >
            <span className="text-xs font-mono text-primary">
              {t('profile.telegram.botLinkLabel')}
            </span>
            <span className="inline-flex items-center gap-1 text-xs font-semibold text-[#229ED9]">
              {t('profile.telegram.botLinkOpen', { bot: telegramBotLabel })}
              <ExternalLink size={13} />
            </span>
          </a>
        ) : (
          <p className="mb-5 text-xs text-muted-foreground">
            {t('profile.telegram.botLinkMissing')}
          </p>
        )}

        <div className="mb-5 rounded-xl border border-border bg-background px-4 py-3">
          <p className="text-[11px] font-mono uppercase tracking-widest text-muted-foreground">
            {t('profile.telegram.statusTitle')}
          </p>
          {isTelegramStatusLoading ? (
            <p className="mt-1 text-xs text-muted-foreground">{t('profile.telegram.statusLoading')}</p>
          ) : telegramIdentity ? (
            <div className="mt-2 space-y-1 text-xs text-primary">
              <p className="font-semibold text-emerald-600">{t('profile.telegram.statusLinked')}</p>
              <p>{t('profile.telegram.statusLinkedAs', { value: telegramIdentityLabel })}</p>
              <p className="text-muted-foreground">{t('profile.telegram.statusLinkedAt', { date: telegramLinkedAtLabel })}</p>
            </div>
          ) : (
            <p className="mt-1 text-xs text-muted-foreground">{t('profile.telegram.statusNotLinked')}</p>
          )}
        </div>

        <div className="flex flex-wrap gap-3">
          <Button
            onClick={handleGenerateTelegramCode}
            disabled={isGeneratingCode}
            className="min-w-[220px] gap-2"
          >
            {isGeneratingCode ? <LoaderCircle size={16} className="animate-spin" /> : null}
            {isGeneratingCode
              ? t('profile.telegram.generating')
              : telegramIdentity
                ? t('profile.telegram.relinkCode')
                : t('profile.telegram.generateCode')}
          </Button>

          <Button
            variant="outline"
            onClick={handleCopyTelegramCommand}
            disabled={!telegramCode?.code || isCopyingCode}
            className="gap-2"
          >
            {isCopyingCode ? <LoaderCircle size={16} className="animate-spin" /> : <Copy size={15} />}
            {t('profile.telegram.copyCommand')}
          </Button>

          {telegramIdentity ? (
            <Button
              variant="outline"
              onClick={handleUnlinkTelegram}
              disabled={isUnlinkingTelegram}
              className="gap-2"
            >
              {isUnlinkingTelegram ? <LoaderCircle size={16} className="animate-spin" /> : null}
              {isUnlinkingTelegram ? t('profile.telegram.unlinking') : t('profile.telegram.unlink')}
            </Button>
          ) : null}
        </div>

        {telegramCode ? (
          <div className="mt-4 rounded-xl border border-border bg-background p-4">
            <p className="text-[11px] font-mono uppercase tracking-widest text-muted-foreground">
              {t('profile.telegram.codeLabel')}
            </p>
            <p className="mt-1 text-xl font-bold tracking-[0.22em] text-primary">{telegramCode.code}</p>
            <p className="mt-2 text-xs text-muted-foreground">
              {t('profile.telegram.expiresAt', { date: telegramExpiresLabel })}
            </p>
            <p className="mt-2 rounded-lg bg-surface px-3 py-2 font-mono text-xs text-accent">
              {t('profile.telegram.commandHint', { command: `/link ${telegramCode.code}` })}
            </p>
          </div>
        ) : null}

        {telegramNotice ? (
          <div
            className={`mt-4 rounded-xl border px-4 py-3 text-sm ${
              telegramNotice.tone === 'success'
                ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700'
                : 'border-rose-500/30 bg-rose-500/10 text-rose-700'
            }`}
          >
            {telegramNotice.text}
          </div>
        ) : null}
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
