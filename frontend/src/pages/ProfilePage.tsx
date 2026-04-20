import React from 'react';
import { useAuth } from '../context/AuthContext';
import { AppLanguage, useLanguage } from '../context/LanguageContext';
import { AppTheme, useTheme } from '../context/ThemeContext';
import { Card } from '../components/ui/Card';
import { ToggleSwitch } from '../components/ui/ToggleSwitch';
import { Globe, Palette, UserRound, Bell, Volume2, Loader, ExternalLink } from 'lucide-react';
import { useI18n } from '../lib/i18n';
import { notificationsEnabled, requestNotificationPermission } from '../lib/notifications';
import { playSuccessMelody, playSound, fetchSoundPreferences } from '../lib/sounds';
import { api } from '../api';

const languageOptions: Array<{ value: AppLanguage; label: string }> = [
  { value: 'en', label: 'English' },
  { value: 'ru', label: 'Русский' },
  { value: 'kk', label: 'Қазақша' },
];

export const ProfilePage: React.FC = () => {
  const { user } = useAuth();
  const { language, setLanguage } = useLanguage();
  const { theme, setTheme } = useTheme();
  const { t } = useI18n();
  
  const [soundsEnabled, setSoundsEnabled] = React.useState(() => {
    if (typeof window === 'undefined') return true;
    return localStorage.getItem('sounds-enabled') !== 'false';
  });
  
  const [notificationsEnabledState, setNotificationsEnabledState] = React.useState(() => {
    if (typeof window === 'undefined') return false;
    const stored = localStorage.getItem('notifications-enabled');
    if (stored !== null) return stored !== 'false';
    return notificationsEnabled();
  });
  const [soundVolume, setSoundVolume] = React.useState(0.7);
  const [isLoadingSounds, setIsLoadingSounds] = React.useState(true);

  // Load sound preferences on mount
  React.useEffect(() => {
    const loadPreferences = async () => {
      try {
        setIsLoadingSounds(true);
        const pref = await fetchSoundPreferences();
        setSoundVolume(pref.volume);
      } catch (err) {
        console.error('Failed to load sound preferences:', err);
      } finally {
        setIsLoadingSounds(false);
      }
    };
    loadPreferences();
  }, []);

  const themeOptions: Array<{ value: AppTheme; label: string }> = [
    { value: 'midnight', label: t('profile.theme.option.midnight') },
    { value: 'ivory', label: t('profile.theme.option.ivory') },
    { value: 'forest', label: t('profile.theme.option.forest') },
    { value: 'ocean', label: t('profile.theme.option.ocean') },
  ];

  const handleToggleSounds = async () => {
    const newValue = !soundsEnabled;
    setSoundsEnabled(newValue);
    localStorage.setItem('sounds-enabled', String(newValue));
    
    // Update API
    try {
      await api.notifications.updateSounds({
        sounds_enabled: newValue,
        volume: soundVolume,
      });
    } catch (err) {
      console.error('Failed to update sound preferences:', err);
    }

    if (newValue) {
      playSuccessMelody(soundVolume);
    }
  };

  const handleVolumeChange = async (newVolume: number) => {
    setSoundVolume(newVolume);
    
    // Update API with debounce
    try {
      await api.notifications.updateSounds({
        volume: newVolume,
        sounds_enabled: soundsEnabled,
      });
    } catch (err) {
      console.error('Failed to update volume:', err);
    }
  };

  const handleTestSound = async () => {
    try {
      await playSound('session-complete', soundVolume);
    } catch (err) {
      console.error('Failed to play test sound:', err);
    }
  };

  const handleToggleNotifications = async () => {
    if (notificationsEnabledState) {
      // Отключить уведомления
      setNotificationsEnabledState(false);
      localStorage.setItem('notifications-enabled', 'false');
    } else {
      // Включить уведомления
      const granted = await requestNotificationPermission();
      setNotificationsEnabledState(granted);
      localStorage.setItem('notifications-enabled', String(granted));
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

      <a 
        href="https://t.me/mathalama_hub" 
        target="_blank" 
        rel="noopener noreferrer"
        className="block group"
      >
        <Card className="border-[#24A1DE]/30 bg-[#24A1DE]/5 p-6 shadow-none transition-colors group-hover:bg-[#24A1DE]/10">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#24A1DE] text-white shadow-lg shadow-[#24A1DE]/20">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="m22 2-7 20-4-9-9-4Z" />
                  <path d="M22 2 11 13" />
                </svg>
              </div>
              <div>
                <h3 className="font-mono text-sm font-bold uppercase tracking-wide text-primary">{t('profile.telegram.title')}</h3>
                <p className="text-xs text-muted-foreground">{t('profile.telegram.subscribe')}</p>
              </div>
            </div>
            <ExternalLink size={16} className="text-[#24A1DE] opacity-50 group-hover:opacity-100 transition-opacity" />
          </div>
        </Card>
      </a>

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

        <div className="mt-6 mb-4 flex items-center gap-2">
          <Palette size={16} className="text-accent" />
          <h2 className="font-mono text-sm font-bold uppercase tracking-wide text-primary">{t('profile.theme')}</h2>
        </div>

        <label className="mb-2 block text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
          {t('profile.interfaceTheme')}
        </label>
        <select
          value={theme}
          onChange={(e) => setTheme(e.target.value as AppTheme)}
          className="h-11 w-full rounded-xl border-2 border-border bg-background px-4 text-sm text-primary outline-none transition-all focus:border-primary/20"
        >
          {themeOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        <p className="mt-3 text-xs text-muted-foreground">
          {t('themeRollout')}
        </p>
      </Card>

      <Card className="border-border bg-surface p-6 shadow-none">
        <div className="mb-4 flex items-center gap-2">
          <Volume2 size={16} className="text-accent" />
          <h2 className="font-mono text-sm font-bold uppercase tracking-wide text-primary">{t('profile.sounds')}</h2>
        </div>

        <label className="mb-4 block text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
          {t('profile.soundsDescription')}
        </label>
        
        <ToggleSwitch
          enabled={soundsEnabled}
          onChange={handleToggleSounds}
          label={soundsEnabled ? t('profile.enabled') : t('profile.disabled')}
        />

        {isLoadingSounds ? (
          <div className="mt-4 flex items-center gap-2 text-xs text-muted-foreground">
            <Loader size={14} className="animate-spin" />
            Loading sound settings...
          </div>
        ) : (
          <div className="mt-4 space-y-4">
            <div>
              <label className="mb-2 flex items-center justify-between">
                <span className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
                  Volume: {Math.round(soundVolume * 100)}%
                </span>
              </label>
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={soundVolume}
                onChange={(e) => handleVolumeChange(parseFloat(e.target.value))}
                disabled={!soundsEnabled}
                className="w-full cursor-pointer accent-accent disabled:opacity-50"
              />
            </div>

            <button
              onClick={handleTestSound}
              disabled={!soundsEnabled}
              className="w-full rounded-lg border border-accent/30 bg-accent/10 px-4 py-2 text-xs font-mono font-bold uppercase text-accent transition-colors hover:bg-accent/20 disabled:opacity-50"
            >
              Test Sound
            </button>
          </div>
        )}
        
        <p className="mt-4 text-xs text-muted-foreground">
          {t('profile.soundsHint')}
        </p>
      </Card>

      <Card className="border-border bg-surface p-6 shadow-none">
        <div className="mb-4 flex items-center gap-2">
          <Bell size={16} className="text-accent" />
          <h2 className="font-mono text-sm font-bold uppercase tracking-wide text-primary">{t('profile.notifications')}</h2>
        </div>

        <label className="mb-4 block text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
          {t('profile.notificationsDescription')}
        </label>
        
        <ToggleSwitch
          enabled={notificationsEnabledState}
          onChange={handleToggleNotifications}
          label={notificationsEnabledState ? t('profile.enabled') : t('profile.disabled')}
        />
        
        <p className="mt-4 text-xs text-muted-foreground">
          {t('profile.notificationsHint')}
        </p>
      </Card>
    </div>
  );
};
