class WebSocketClient {
    constructor() {
        this.ws = null;
        this.url = '';
        this.handlers = new Map();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.heartbeatTimer = null;
        this.reconnectTimer = null;
        this.status = 'disconnected';
    }
    connect(token) {
        // 默认连接到 tasks 通道（个人通知）
        this.url = `/ws/tasks?token=${token}&t=${Date.now()}`;
        this.connectWs();
    }
    connectWs() {
        this.status = 'connecting';
        try {
            this.ws = new WebSocket(this.url);
            this.ws.onopen = () => {
                this.status = 'connected';
                this.reconnectAttempts = 0;
                this.startHeartbeat();
                this.notifyStatusChange();
            };
            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    if (message.event_type === 'pong')
                        return;
                    this.dispatchMessage(message);
                }
                catch { /* ignored */ }
            };
            this.ws.onclose = () => {
                this.status = 'disconnected';
                this.stopHeartbeat();
                this.notifyStatusChange();
                this.scheduleReconnect();
            };
            this.ws.onerror = () => {
                this.status = 'disconnected';
                this.notifyStatusChange();
            };
        }
        catch {
            this.scheduleReconnect();
        }
    }
    dispatchMessage(message) {
        const globalHandlers = this.handlers.get('*') || new Set();
        const channelHandlers = this.handlers.get(message.event_type) || new Set();
        for (const h of globalHandlers)
            h(message);
        for (const h of channelHandlers)
            h(message);
    }
    subscribe(eventType, handler) {
        if (!this.handlers.has(eventType))
            this.handlers.set(eventType, new Set());
        this.handlers.get(eventType).add(handler);
    }
    unsubscribe(eventType, handler) {
        const h = this.handlers.get(eventType);
        if (h)
            h.delete(handler);
    }
    send(data) {
        if (this.ws?.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data));
        }
    }
    startHeartbeat() {
        this.heartbeatTimer = setInterval(() => {
            if (this.ws?.readyState === WebSocket.OPEN) {
                this.send({ event_type: 'ping', timestamp: Date.now() });
            }
        }, 30000);
    }
    stopHeartbeat() {
        if (this.heartbeatTimer)
            clearInterval(this.heartbeatTimer);
        this.heartbeatTimer = null;
    }
    scheduleReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts)
            return;
        const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 16000);
        this.reconnectAttempts++;
        this.reconnectTimer = setTimeout(() => {
            this.connectWs();
        }, delay);
    }
    notifyStatusChange() {
        const handlers = this.handlers.get('connection_status') || new Set();
        for (const h of handlers)
            h({ event_type: 'connection_status', payload: { status: this.status }, timestamp: Date.now() });
    }
    disconnect() {
        this.stopHeartbeat();
        if (this.reconnectTimer)
            clearTimeout(this.reconnectTimer);
        this.ws?.close();
        this.ws = null;
        this.status = 'disconnected';
    }
}
export const wsClient = new WebSocketClient();
