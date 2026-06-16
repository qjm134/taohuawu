package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/watertown/guide/internal/agent"
	"github.com/watertown/guide/internal/database"
	"github.com/watertown/guide/internal/emotion"
	"github.com/watertown/guide/internal/knowledge"
	"github.com/watertown/guide/internal/llm"
	"github.com/watertown/guide/internal/websocket"
	"github.com/watertown/guide/pkg/logging"
	"github.com/watertown/guide/pkg/utils"
)

// WebSocketHandler WebSocket 处理器
type WebSocketHandler struct {
	hub            *websocket.Hub
	sessionManager *agent.SessionManager
	runtime        *agent.Runtime
	playerRepo     database.PlayerRepository
	convRepo       database.ConversationRepository
	auditRepo      database.AuditRepository
	logger         logging.Logger
}

// NewWebSocketHandler 创建 WebSocket 处理器
func NewWebSocketHandler(
	hub *websocket.Hub,
	sessionManager *agent.SessionManager,
	runtime *agent.Runtime,
	playerRepo database.PlayerRepository,
	convRepo database.ConversationRepository,
	auditRepo database.AuditRepository,
	logger logging.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:            hub,
		sessionManager: sessionManager,
		runtime:        runtime,
		playerRepo:     playerRepo,
		convRepo:       convRepo,
		auditRepo:      auditRepo,
		logger:         logger,
	}
}

// Handle 处理 WebSocket 连接
func (h *WebSocketHandler) Handle(c *gin.Context) {
	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", "error", err)
		return
	}

	// 创建客户端
	client := websocket.NewClient(conn, "", "", nil)

	// 注册客户端到 Hub
	h.hub.register <- client

	// 启动读写泵
	go client.WritePump()
	go client.ReadPump(h.hub, h.handleMessage)
}

// handleMessage 处理消息
func (h *WebSocketHandler) handleMessage(client *websocket.Client, message []byte) {
	// 解析消息
	var msg websocket.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		h.logger.Error("Failed to parse message", "error", err)
		return
	}

	switch msg.Type {
	case websocket.MessageTypeConnection:
		h.handleConnection(client, &msg)
	case websocket.MessageTypeChatMessage:
		h.handleChatMessage(client, &msg)
	case websocket.MessageTypePing:
		h.handlePing(client, &msg)
	default:
		h.logger.Warn("Unknown message type", "type", msg.Type)
	}
}

// handleConnection 处理连接消息
func (h *WebSocketHandler) handleConnection(client *websocket.Client, msg *websocket.Message) {
	var payload websocket.ConnectionPayload
	if err := msg.ParsePayload(&payload); err != nil {
		h.logger.Error("Failed to parse connection payload", "error", err)
		return
	}

	// 设置客户端信息
	client.TenantID = msg.TenantID
	client.PlayerID = payload.PlayerID

	// 查找或创建玩家
	player, err := h.playerRepo.GetByDeviceID(payload.DeviceID, msg.TenantID)
	if err != nil {
		// 创建新玩家
		player = &database.Player{
			ID:              uuid.New().String(),
			TenantID:        msg.TenantID,
			Nickname:        payload.Nickname,
			DeviceID:        payload.DeviceID,
			FirstVisitTime:  time.Now(),
			LastVisitTime:   time.Now(),
			TotalDialogues:  0,
		}
		if err := h.playerRepo.Create(player); err != nil {
			h.logger.Error("Failed to create player", "error", err)
			return
		}
	} else {
		// 更新最后访问时间
		_ = h.playerRepo.UpdateLastVisit(player.ID)
	}

	// 获取会话
	session := h.runtime.GetSession(player.ID, msg.TenantID)
	session.Nickname = payload.Nickname

	// 处理欢迎
	reply, err := h.runtime.HandleWelcome(context.Background(), session)
	if err != nil {
		h.logger.Error("Failed to handle welcome", "error", err)
		return
	}

	// 标记已访问
	h.runtime.MarkVisited(session.ID)

	// 构建欢迎消息
	welcomeMsg, _ := websocket.NewMessage(
		websocket.MessageTypeWelcome,
		msg.RequestID,
		msg.TenantID,
		websocket.WelcomePayload{
			GuideName:    agent.GuideName,
			Message:      reply,
			IsFirstVisit: session.IsFirstVisit,
			Tips:         []string{"点击输入框与小荷对话", "可以问我关于游戏的问题"},
		},
	)

	_ = client.SendMessage(welcomeMsg)
}

// handleChatMessage 处理聊天消息
func (h *WebSocketHandler) handleChatMessage(client *websocket.Client, msg *websocket.Message) {
	var payload websocket.ChatMessagePayload
	if err := msg.ParsePayload(&payload); err != nil {
		h.logger.Error("Failed to parse chat payload", "error", err)
		return
	}

	// 查找玩家
	player, err := h.playerRepo.GetByID(payload.PlayerID)
	if err != nil {
		h.logger.Error("Player not found", "player_id", payload.PlayerID)
		return
	}

	// 获取会话
	session := h.runtime.GetSession(player.ID, msg.TenantID)

	// 处理聊天
	reply, emotion, err := h.runtime.HandleChat(context.Background(), session, payload.Message)
	if err != nil {
		h.logger.Error("Failed to handle chat", "error", err)

		// 返回错误消息
		errMsg, _ := websocket.NewMessage(
			websocket.MessageTypeError,
			msg.RequestID,
			msg.TenantID,
			websocket.ErrorPayload{
				Code:    "CHAT_ERROR",
				Message: "抱歉，我现在无法回答你的问题。请稍后再试。",
			},
		)
		_ = client.SendMessage(errMsg)
		return
	}

	// 增加对话计数
	_ = h.playerRepo.IncrementDialogues(player.ID)

	// 保存对话记录
	conv := &database.Conversation{
		ID:          uuid.New().String(),
		PlayerID:    player.ID,
		TenantID:    msg.TenantID,
		SessionID:   session.ID,
		UserMessage: payload.Message,
		AIMessage:   reply,
		Emotion:     emotion,
		CreatedAt:   time.Now(),
	}
	_ = h.convRepo.Create(conv)

	// 构建回复消息
	replyMsg, _ := websocket.NewMessage(
		websocket.MessageTypeNPCReply,
		msg.RequestID,
		msg.TenantID,
		websocket.NPCReplyPayload{
			GuideName: agent.GuideName,
			Message:   reply,
			Emotion:   emotion,
			Actions:   []string{},
		},
	)

	_ = client.SendMessage(replyMsg)
}

// handlePing 处理心跳
func (h *WebSocketHandler) handlePing(client *websocket.Client, msg *websocket.Message) {
	pongMsg, _ := websocket.NewMessage(
		websocket.MessageTypePong,
		msg.RequestID,
		msg.TenantID,
		websocket.PongPayload{
			ServerTime: time.Now().UnixMilli(),
		},
	)
	_ = client.SendMessage(pongMsg)
}