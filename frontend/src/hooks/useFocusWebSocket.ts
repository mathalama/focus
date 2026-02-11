import { useEffect, useState, useCallback, useRef } from 'react';

export interface FocusSessionEvent {
  action: 'started' | 'stopped';
  duration: number; // in seconds
  timestamp: string;
  user_id: string;
}

interface WebSocketMessage {
  type: string;
  payload: FocusSessionEvent;
}

export function useFocusWebSocket(token: string | null, apiUrl: string | null) {
  const [focusStatus, setFocusStatus] = useState<FocusSessionEvent | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout>>();

  const connect = useCallback(() => {
    if (!token || !apiUrl) {
      setIsConnected(false);
      return;
    }

    try {
      // Convert HTTP/HTTPS to WS/WSS
      const wsUrl = (apiUrl.startsWith('https')
        ? apiUrl.replace('https', 'wss')
        : apiUrl.startsWith('http')
          ? apiUrl.replace('http', 'ws')
          : apiUrl) + '/api/v1/ws/focus';

      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('Focus WebSocket connected');
        setIsConnected(true);
        
        // Send auth message
        ws.send(
          JSON.stringify({
            type: 'auth',
            payload: { token }
          })
        );
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          
          if (message.type === 'focus_session_event') {
            setFocusStatus(message.payload);
          }
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e);
        }
      };

      ws.onerror = (event) => {
        const errorMsg = `WebSocket error: ${event.type}`;
        console.error(errorMsg);
        setIsConnected(false);
      };

      ws.onclose = () => {
        console.log('Focus WebSocket disconnected');
        setIsConnected(false);
        wsRef.current = null;
        
        // Attempt to reconnect in 5 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, 5000);
      };

      return () => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.close();
        }
      };
    } catch (e) {
      const errorMsg = e instanceof Error ? e.message : 'Unknown error';
      console.error('Failed to create WebSocket:', errorMsg);
      setIsConnected(false);
    }
  }, [token, apiUrl]);

  useEffect(() => {
    const cleanup = connect();
    return () => {
      cleanup?.();
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [connect]);

  return {
    focusStatus,
    isConnected
  };
}
