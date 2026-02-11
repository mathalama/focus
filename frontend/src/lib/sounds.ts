// Sound management utilities
import type { NotificationSound } from '../api/notifications';

type SoundType = 'session-complete' | 'break-end' | 'notification';

const defaultSoundMap: Record<SoundType, string> = {
  'session-complete': '/sounds/session-complete.mp3',
  'break-end': '/sounds/break-end.mp3',
  'notification': '/sounds/notification.mp3',
};

let audioContext: AudioContext | null = null;
let soundPreferences: NotificationSound | null = null;
let soundPreferencesPromise: Promise<NotificationSound> | null = null;

export const initAudioContext = () => {
  if (typeof window === 'undefined') return;
  if (!audioContext && (window.AudioContext || (window as any).webkitAudioContext)) {
    audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
  }
};

/**
 * Fetch user's notification sound preferences from API
 */
export const fetchSoundPreferences = async (): Promise<NotificationSound> => {
  if (soundPreferences) return soundPreferences;
  if (soundPreferencesPromise) return soundPreferencesPromise;

  soundPreferencesPromise = (async () => {
    try {
      const { api } = await import('../api');
      const { sound } = await api.notifications.getSounds();
      soundPreferences = sound;
      return sound;
    } catch (err) {
      console.warn('Failed to fetch sound preferences:', err);
      // Return default preferences on error
      return {
        id: '',
        user_id: '',
        session_complete_sound: defaultSoundMap['session-complete'],
        break_end_sound: defaultSoundMap['break-end'],
        notification_sound: defaultSoundMap['notification'],
        volume: 0.7,
        sounds_enabled: localStorage.getItem('sounds-enabled') !== 'false',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
    }
  })();

  return soundPreferencesPromise;
};

/**
 * Get the sound URL for a specific sound type
 */
export const getSoundUrl = async (soundType: SoundType): Promise<string> => {
  const preferences = await fetchSoundPreferences();
  
  switch (soundType) {
    case 'session-complete':
      return preferences.session_complete_sound;
    case 'break-end':
      return preferences.break_end_sound;
    case 'notification':
      return preferences.notification_sound;
  }
};

/**
 * Play a sound by type (with Web Audio API fallback if file not available)
 */
export const playSound = async (soundType: SoundType, volume?: number): Promise<void> => {
  if (typeof window === 'undefined') return;
  
  try {
    initAudioContext();
    const preferences = await fetchSoundPreferences();
    
    // Check if sounds are enabled
    if (localStorage.getItem('sounds-enabled') === 'false') {
      return;
    }

    const audioPath = await getSoundUrl(soundType);
    const audioVolume = volume ?? preferences.volume;
    
    // Don't play if volume is 0 or negative
    if (audioVolume <= 0) {
      return;
    }
    
    const audio = new Audio(audioPath);
    audio.volume = Math.min(audioVolume, 1);
    
    try {
      await audio.play();
    } catch (playErr) {
      // File not found or play failed - use Web Audio API fallback
      console.warn(`Failed to play sound ${soundType}, using fallback beep`);
      
      switch (soundType) {
        case 'session-complete':
          playSuccessMelody(audioVolume);
          break;
        case 'break-end':
          playBeep(600, 300, audioVolume);
          playBeep(800, 300, audioVolume);
          break;
        case 'notification':
          playBeep(523, 100, audioVolume * 0.8);
          break;
      }
    }
  } catch (err) {
    console.warn(`Sound playback error for ${soundType}:`, err);
  }
};

/**
 * Play a simple beep sound using Web Audio API
 */
export const playBeep = (frequency = 800, duration = 200, volume = 0.3): void => {
  if (typeof window === 'undefined') return;
  
  // Don't play if volume is 0 or negative
  if (volume <= 0) return;
  
  try {
    initAudioContext();
    if (!audioContext) return;

    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();
    
    oscillator.connect(gainNode);
    gainNode.connect(audioContext.destination);
    
    oscillator.frequency.value = frequency;
    gainNode.gain.setValueAtTime(volume, audioContext.currentTime);
    gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + duration / 1000);
    
    oscillator.start(audioContext.currentTime);
    oscillator.stop(audioContext.currentTime + duration / 1000);
  } catch (err) {
    console.warn('Beep sound error:', err);
  }
};

/**
 * Play a success melody (3 beeps ascending)
 */
export const playSuccessMelody = (volume = 0.4): void => {
  if (typeof window === 'undefined') return;
  
  // Don't play if volume is 0 or negative
  if (volume <= 0) return;
  
  try {
    initAudioContext();
    if (!audioContext) return;

    const notes = [
      { freq: 523.25, delay: 0 },    // C5
      { freq: 659.25, delay: 150 },  // E5
      { freq: 783.99, delay: 300 },  // G5
    ];

    notes.forEach(({ freq, delay }) => {
      setTimeout(() => {
        playBeep(freq, 200, volume);
      }, delay);
    });
  } catch (err) {
    console.warn('Success melody error:', err);
  }
};
