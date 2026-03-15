export interface WSMessage {
  type: string;
  payload: any;
}

type MessageHandler = (msg: WSMessage) => void;

export class QuizSocket {
  private ws: WebSocket | null = null;
  private url: string;
  private token: string | null;
  private guestName: string | undefined;
  private handlers: Map<string, MessageHandler[]> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private shouldReconnect = true;
  private wasConnected = false;
  private intentionalClose = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(code: string, token: string | null, guestName?: string) {
    const wsUrl = process.env.REACT_APP_WS_URL || (process.env.NODE_ENV === 'development' ? 'ws://localhost:8080/ws' : `wss://${window.location.host}/ws`);
    this.url = `${wsUrl}/${code}`;
    this.token = token;
    this.guestName = guestName;
  }

  connect(): void {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      // Send auth as first message instead of query params
      const authPayload: Record<string, string> = {};
      if (this.token) {
        authPayload.token = this.token;
      } else {
        authPayload.guest = this.guestName || 'Player';
      }
      this.ws!.send(JSON.stringify({ type: 'auth', payload: authPayload }));

      const isReconnect = this.wasConnected;
      this.reconnectAttempts = 0;
      this.wasConnected = true;

      if (isReconnect) {
        this.emit('reconnected', { message: 'Connection restored' });
      }
    };

    this.ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        this.emit(msg.type, msg.payload);
      } catch (e) {
        console.error('Failed to parse WS message:', e);
      }
    };

    this.ws.onclose = (event) => {
      if (this.intentionalClose) return;
      if (!this.wasConnected) {
        let message = 'Could not reach the quiz server. It may be offline or starting up — please try again in a moment.';
        if (event.code === 1008) {
          message = 'Authentication failed. Please log in again.';
        } else if (event.code === 1003) {
          message = 'Username is already taken in this quiz. Try a different name.';
        }
        this.emit('connection_error', { message, code: event.code });
        return;
      }

      this.emit('disconnected', { message: 'Connection lost' });

      if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
        const delay = Math.pow(2, this.reconnectAttempts) * 1000;
        this.reconnectAttempts++;
        this.emit('reconnecting', { attempt: this.reconnectAttempts, maxAttempts: this.maxReconnectAttempts, delay });
        this.reconnectTimer = setTimeout(() => this.connect(), delay);
      } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
        this.emit('connection_error', { message: 'Unable to reconnect. Please refresh the page.' });
      }
    };

    this.ws.onerror = (err) => {
      console.error('WebSocket error:', err);
    };
  }

  private emit(type: string, payload: any): void {
    const msg: WSMessage = { type, payload };
    const handlers = this.handlers.get(type) || [];
    handlers.forEach((h) => h(msg));

    const allHandlers = this.handlers.get('*') || [];
    allHandlers.forEach((h) => h(msg));
  }

  on(type: string, handler: MessageHandler): () => void {
    const existing = this.handlers.get(type) || [];
    existing.push(handler);
    this.handlers.set(type, existing);

    return () => {
      const handlers = this.handlers.get(type) || [];
      this.handlers.set(type, handlers.filter((h) => h !== handler));
    };
  }

  send(type: string, payload?: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    }
  }

  disconnect(): void {
    this.intentionalClose = true;
    this.shouldReconnect = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.ws?.close();
    this.ws = null;
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}
