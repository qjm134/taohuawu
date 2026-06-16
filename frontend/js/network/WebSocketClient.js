// WebSocket 客户端
class WebSocketClient {
    constructor(url, options = {}) {
        this.url = url;
        this.reconnectInterval = options.reconnectInterval || 3000;
        this.pingInterval = options.pingInterval || 30000;
        this.socket = null;
        this.messageQueue = [];
        this.isConnected = false;
        this.pingTimer = null;
        this.reconnectTimer = null;
        this.requestId = 0;

        // 事件处理器
        this.onMessage = null;
        this.onConnect = null;
        this.onDisconnect = null;
        this.onError = null;
    }

    connect() {
        try {
            this.socket = new WebSocket(this.url);

            this.socket.onopen = () => {
                this.isConnected = true;
                this.startPing();
                this.processMessageQueue();
                if (this.onConnect) {
                    this.onConnect();
                }
                console.log('WebSocket connected');
            };

            this.socket.onmessage = (event) => {
                this.handleMessage(event.data);
            };

            this.socket.onerror = (error) => {
                console.error('WebSocket error:', error);
                if (this.onError) {
                    this.onError(error);
                }
            };

            this.socket.onclose = () => {
                this.isConnected = false;
                this.stopPing();
                this.scheduleReconnect();
                if (this.onDisconnect) {
                    this.onDisconnect();
                }
                console.log('WebSocket disconnected');
            };
        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
            this.scheduleReconnect();
        }
    }

    disconnect() {
        this.stopPing();
        this.stopReconnect();
        if (this.socket) {
            this.socket.close();
            this.socket = null;
        }
        this.isConnected = false;
    }

    send(message) {
        if (!this.isConnected) {
            this.messageQueue.push(message);
            console.log('Message queued (not connected)');
            return;
        }

        try {
            const data = JSON.stringify(message);
            this.socket.send(data);
        } catch (error) {
            console.error('Failed to send message:', error);
            this.messageQueue.push(message);
        }
    }

    sendConnection(playerId, nickname, deviceId) {
        const message = {
            type: MESSAGE_TYPES.CONNECTION,
            requestId: this.getRequestId(),
            tenantId: 'tenant_001',
            timestamp: Date.now(),
            payload: {
                playerId: playerId,
                nickname: nickname,
                deviceId: deviceId,
            },
        };
        this.send(message);
    }

    sendChatMessage(message, playerId) {
        const msg = {
            type: MESSAGE_TYPES.CHAT_MESSAGE,
            requestId: this.getRequestId(),
            tenantId: 'tenant_001',
            timestamp: Date.now(),
            payload: {
                message: message,
                playerId: playerId,
            },
        };
        this.send(msg);
    }

    sendPing() {
        const message = {
            type: MESSAGE_TYPES.PING,
            requestId: this.getRequestId(),
            tenantId: 'tenant_001',
            timestamp: Date.now(),
            payload: {},
        };
        this.send(message);
    }

    handleMessage(data) {
        try {
            const message = JSON.parse(data);

            // 处理 PONG
            if (message.type === MESSAGE_TYPES.PONG) {
                return;
            }

            if (this.onMessage) {
                this.onMessage(message);
            }
        } catch (error) {
            console.error('Failed to handle message:', error);
        }
    }

    processMessageQueue() {
        while (this.messageQueue.length > 0) {
            const message = this.messageQueue.shift();
            this.send(message);
        }
    }

    startPing() {
        this.stopPing();
        this.pingTimer = setInterval(() => {
            this.sendPing();
        }, this.pingInterval);
    }

    stopPing() {
        if (this.pingTimer) {
            clearInterval(this.pingTimer);
            this.pingTimer = null;
        }
    }

    scheduleReconnect() {
        this.stopReconnect();
        this.reconnectTimer = setTimeout(() => {
            this.connect();
        }, this.reconnectInterval);
    }

    stopReconnect() {
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }
    }

    getRequestId() {
        return `req_${Date.now()}_${this.requestId++}`;
    }

    generateDeviceId() {
        return 'device_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    generatePlayerId() {
        return 'player_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
}