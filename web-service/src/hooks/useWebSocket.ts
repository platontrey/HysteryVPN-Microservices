import { useEffect, useState, useRef, useCallback } from 'react';

export interface WSMessage {
  type: 'traffic_update' | 'user_status' | 'device_online' | 'error';
  user_id?: string;
  data?: any;
  timestamp: string;
}

export interface TrafficUpdate {
  id: string;
  user_id: string;
  device_id?: string;
  upload: number;
  download: number;
  total: number;
  recorded_at: string;
  created_at: string;
}

export const useWebSocket = (url: string, token: string) => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);
  const [trafficUpdates, setTrafficUpdates] = useState<TrafficUpdate[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<string>('disconnected');
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const reconnectDelay = 3000;

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    try {
      // Use secure WebSocket in production
      const protocol = url.startsWith('wss') ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${url.replace(/^wss?:\/\//, '')}?token=${encodeURIComponent(token)}`;
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        setConnectionStatus('connected');
        reconnectAttemptsRef.current = 0;

        // Send initial ping
        sendMessage({ type: 'ping' });
      };

      ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          setLastMessage(message);

          // Handle different message types
          switch (message.type) {
            case 'traffic_update':
              if (message.data) {
                setTrafficUpdates(prev => [message.data, ...prev.slice(0, 9)]); // Keep last 10 updates
              }
              break;
            case 'user_status':
              setConnectionStatus(message.data?.status || 'unknown');
              break;
            case 'device_online':
              // Handle device status updates
              break;
            case 'error':
              console.error('WebSocket error:', message.data);
              break;
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        setIsConnected(false);
        setConnectionStatus('disconnected');
        wsRef.current = null;

        // Attempt to reconnect if not a normal closure
        if (event.code !== 1000 && reconnectAttemptsRef.current < maxReconnectAttempts) {
          reconnectAttemptsRef.current++;
          console.log(`Attempting to reconnect (${reconnectAttemptsRef.current}/${maxReconnectAttempts})...`);
          reconnectTimeoutRef.current = setTimeout(connect, reconnectDelay);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionStatus('error');
      };

      wsRef.current = ws;
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      setConnectionStatus('error');
    }
  }, [url, token]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }

    if (wsRef.current) {
      wsRef.current.close(1000, 'Client disconnect');
      wsRef.current = null;
    }

    setIsConnected(false);
    setConnectionStatus('disconnected');
    reconnectAttemptsRef.current = 0;
  }, []);

  const sendMessage = useCallback((message: Partial<WSMessage>) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      const fullMessage: WSMessage = {
        ...message,
        timestamp: new Date().toISOString(),
      };
      wsRef.current.send(JSON.stringify(fullMessage));
    } else {
      console.warn('WebSocket is not connected, cannot send message');
    }
  }, []);

  const subscribeToTraffic = useCallback(() => {
    sendMessage({ type: 'subscribe_traffic' });
  }, [sendMessage]);

  // Connect on mount and token change
  useEffect(() => {
    if (token) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [connect, disconnect, token]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    lastMessage,
    trafficUpdates,
    connectionStatus,
    connect,
    disconnect,
    sendMessage,
    subscribeToTraffic,
  };
};