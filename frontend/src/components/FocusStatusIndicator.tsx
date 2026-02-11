import { useAuth } from '../context/AuthContext';
import { useFocusWebSocket } from '../hooks/useFocusWebSocket';
import { useEffect, useState } from 'react';

export function FocusStatusIndicator() {
  const { token } = useAuth();
  const apiUrl = localStorage.getItem('focus_api_url') || window.location.origin;
  
  const { focusStatus } = useFocusWebSocket(token, apiUrl);
  const [elapsedSeconds, setElapsedSeconds] = useState(0);

  useEffect(() => {
    if (!focusStatus || focusStatus.action !== 'started') {
      setElapsedSeconds(0);
      return;
    }

    const startTime = new Date(focusStatus.timestamp).getTime();
    const interval = setInterval(() => {
      const now = Date.now();
      const elapsed = Math.floor((now - startTime) / 1000);
      setElapsedSeconds(elapsed);
    }, 1000);

    return () => clearInterval(interval);
  }, [focusStatus]);

  if (!focusStatus || focusStatus.action !== 'started') {
    return null;
  }

  const minutes = Math.floor(elapsedSeconds / 60);
  const seconds = elapsedSeconds % 60;

  return (
    <div className="fixed bottom-4 right-4 bg-gradient-to-r from-purple-500 to-pink-500 text-white px-4 py-3 rounded-lg shadow-lg">
      <div className="flex items-center gap-3">
        <div className="w-3 h-3 bg-white rounded-full animate-pulse"></div>
        <div>
          <p className="text-sm font-semibold">Focus mode active</p>
          <p className="text-xs opacity-90">
            {minutes.toString().padStart(2, '0')}:{seconds.toString().padStart(2, '0')} elapsed
          </p>
        </div>
      </div>
    </div>
  );
}
