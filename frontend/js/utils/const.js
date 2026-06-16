// 游戏常量
const GAME_CONFIG = {
    width: 1200,
    height: 800,
    backgroundColor: '#87CEEB',
    tileSize: 64,
};

// NPC 配置
const NPC_CONFIG = {
    guide: {
        name: '小荷',
        position: { x: 600, y: 400 },
        color: 0xFF69B4,
        scale: 1.5,
    },
};

// WebSocket 配置
const WS_CONFIG = {
    url: 'ws://localhost:8080/ws/game',
    reconnectInterval: 3000,
    pingInterval: 30000,
};

// 消息类型
const MESSAGE_TYPES = {
    CONNECTION: 'CONNECTION',
    CHAT_MESSAGE: 'CHAT_MESSAGE',
    PING: 'PING',
    WELCOME: 'WELCOME',
    NPC_REPLY: 'NPC_REPLY',
    ERROR: 'ERROR',
    PONG: 'PONG',
};

// 表情符号
const EMOJIS = {
    happy: '😊',
    sad: '😢',
    angry: '😠',
    confused: '😕',
    excited: '🎉',
    neutral: '😊',
};

// 颜色配置
const COLORS = {
    dialogBox: 0xFFFFFF,
    dialogBorder: 0x4A4A4A,
    playerName: 0x333333,
    npcName: 0xFF69B4,
    dialogText: 0x333333,
    inputBox: 0xFFFFFF,
    inputBorder: 0xCCCCCC,
    inputText: 0x333333,
    inputPlaceholder: 0x999999,
};