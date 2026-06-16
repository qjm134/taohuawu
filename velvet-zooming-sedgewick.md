# 江南水乡智能导游系统 - 实施计划

## Context

创建一个基于2D在线游戏的智能导游系统，玩家进入江南水乡场景后，由NPC少女导游提供游戏指引和基础问答服务。系统需要支持上下文记忆、情绪感知、成本控制和资源隔离等企业级特性。

---

## 一、架构设计

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端浏览器                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Phaser 3 游戏引擎 (Canvas渲染)                          │   │
│  │  - 场景层 (SceneLayer)                                   │   │
│  │  - 角色层 (CharacterLayer: NPC少女)                      │   │
│  │  - UI层 (UILayer: 对话框)                                │   │
│  │  - WebSocketClient (实时通信)                            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                        ↓ WebSocket                              │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                       后端服务 (Go 1.22)                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  HTTP Server (Gin框架)                                    │   │
│  │  - REST API: /health, /metrics, /audit                   │   │
│  │  - WebSocket Handler: /ws/game                           │   │
│  └──────────────────────────────────────────────────────────┘   │
│                        ↓                                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  WebSocket Manager                                         │   │
│  │  - 连接管理 (ConnectionPool)                               │   │
│  │  - 消息路由 (MessageRouter)                                │   │
│  │  - 租户隔离 (TenantPool隔离)                               │   │
│  └──────────────────────────────────────────────────────────┘   │
│                        ↓                                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Agent Runtime (导游 Agent核心)                            │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  Session Manager                                     │ │   │
│  │  │  - 会话状态管理                                      │ │   │
│  │  │  - 消息历史 (带摘要压缩)                             │ │   │
│  │  │  - 上下文缓存 (Redis可选)                            │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  LLM Adapter                                         │ │   │
│  │  │  - GLM-4.7 调用 (主模型)                             │ │   │
│  │  │  - 备用模型 (熔断降级)                               │ │   │
│  │  │  - 超时控制 (10s)                                    │ │   │
│  │  │  - 重试机制 (指数退避)                               │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  Tool Registry (工具注册表)                          │ │   │
│  │  │  - get_player_info()                                 │ │   │
│  │  │  - get_game_guide()                                  │ │   │
│  │  │  - detect_emotion()                                  │ │   │
│  │  │  - 超时控制 (5s)                                     │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  Cost Optimizer                                      │ │   │
│  │  │  - 相似问题缓存 (Embedding >0.95)                    │ │   │
│  │  │  - 历史消息摘要                                       │ │   │
│  │  │  - LLM调用次数统计                                   │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  Circuit Breaker (熔断器)                            │ │   │
│  │  │  - 失败率监控                                         │ │   │
│  │  │  - 开启状态切换                                       │ │   │
│  │  │  - 半开启探测                                         │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────┘   │
│                        ↓                                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Data Layer                                               │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  PostgreSQL                                         │ │   │
│  │  │  - players (玩家表)                                 │ │   │
│  │  │  - conversations (对话记录表)                       │ │   │
│  │  │  - audit_logs (审计日志表)                          │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  JSON Knowledge Base                                │ │   │
│  │  │  - game_faq.json (游戏FAQ)                          │ │   │
│  │  │  - game_rules.json (游戏规则)                       │ │   │
│  │  │  - scenario_desc.json (场景描述)                     │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────┘   │
│                        ↓                                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Observability                                            │   │
│  │  - OpenTelemetry Tracing (请求链路追踪)                   │   │
│  │  - Prometheus Metrics (成本/性能指标)                     │   │
│  │  - Audit Logging (对话审计)                               │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 核心流程图

```
玩家进入游戏流程:
┌─────────┐      1. WebSocket连接       ┌──────────────────┐
│ 客户端  │ ─────────────────────────────► │  WebSocket       │
│ Phaser  │                             │  Handler          │
│ 场景    │      2. 创建会话             │                  │
└─────────┘ ◄──────────────────────────── │  SessionMgr      │
              3. 返回欢迎消息            └──────────────────┘
                    (NPC少女自动问候)              │
                                                  │
                                                  ▼
                                         ┌──────────────────┐
                                         │  Agent Runtime   │
                                         │  (欢迎接待能力)   │
                                         └──────────────────┘

对话交互流程:
┌─────────┐      1. 玩家输入             ┌──────────────────┐
│ 玩家    │ ─────────────────────────────► │  WebSocket       │
│ 输入框  │                             │  Handler          │
└─────────┘      2. 识别租户             │                  │
                  (tenant_id)             │  MessageRouter   │
                                         └──────────────────┘
                                                  │
                                                  ▼
                                         ┌──────────────────┐
                                         │  Cost Optimizer   │
                                         │  (缓存检查)       │
                                         └──────────────────┘
                                         │ 命中?              │
                    是                   │     │ 否
◄─────────────────────────────────────────┘     │
返回缓存答案                                      ▼
                                          ┌──────────────────┐
                                          │  Circuit Breaker │
                                          │  (熔断检查)       │
                                          └──────────────────┘
                                          │ 开启?             │
                    否                    │     │ 是
◄─────────────────────────────────────────┘     │
调用备用/兜底                                      ▼
                                          ┌──────────────────┐
                                          │  Agent Runtime   │
                                          │  - 情绪感知       │
                                          │  - 工具调用       │
                                          │  - LLM生成        │
                                          └──────────────────┘
                                          │                  │
                                          ▼                  ▼
                                   ┌──────────┐    ┌──────────┐
                                   │ Tool     │    │ LLM      │
                                   │ 5s超时   │    │ 10s超时  │
                                   └──────────┘    └──────────┘
                                          │                  │
                                          └────────┬─────────┘
                                                   ▼
                                          ┌──────────────────┐
                                          │  摘要压缩         │
                                          │  (历史消息)       │
                                          └──────────────────┘
                                                   │
                                                   ▼
                                          ┌──────────────────┐
                                          │  审计日志         │
                                          │  PostgreSQL       │
                                          └──────────────────┘
                                                   │
◄──────────────────────────────────────────────────┘
        返回NPC回复 (带情绪调整)
```

### 1.3 时序图

```
玩家      Phaser      WebSocket      SessionMgr     Agent      Tool       LLM      PG
 │           │              │              │          │          │          │         │
 ├─连接──────►│              │              │          │          │          │         │
 │           ├─建立WS连接───►│              │          │          │          │         │
 │           │              ├─创建会话─────►│          │          │          │         │
 │           │              │              ├─检查首次──►│          │          │         │
 │           │              │              │◄─返回true─┤          │          │         │
 │           │              │              ├─欢迎处理──►│          │          │         │
 │           │              │              │          ├─调用LLM────────────────►│       │
 │           │              │              │◄─返回欢迎─┤◄────────────┤          │       │
 │           │              │◄─推送欢迎─────┤          │          │          │         │
 │◄─显示NPC问候─────────────┤              │          │          │          │         │
 │           │              │              │          │          │          │         │
 ├─输入──────►│              │              │          │          │          │         │
 │           ├─发送消息─────►│              │          │          │          │         │
 │           │              ├─消息路由─────►│          │          │          │         │
 │           │              │              ├─成本检查─►│          │          │         │
 │           │              │              │◄─需调用──┤          │          │         │
 │           │              │              ├─处理请求─►│          │          │         │
 │           │              │              │          ├─情绪感知─┤          │         │
 │           │              │              │          ├─工具调用────────────►│       │
 │           │              │              │          │◄─返回数据─┤◄─────────┤         │
 │           │              │              │          ├─调用LLM────────────────►│       │
 │           │              │              │          │◄─生成回复─┤◄────────────┤         │
 │           │              │              │◄─返回回复─┤          │          │         │
 │           │              │              ├─写审计日志────────────────────────────►│
 │           │              │◄─推送NPC回复─┤          │          │          │         │
 │◄─显示对话框─────────────────────────────┤              │          │          │         │
 │           │              │              │          │          │          │         │
```

---

## 二、模块拆分

### 2.1 后端模块 (Go)

```
backend/
├── cmd/
│   └── server/
│       └── main.go                    # 程序入口
│
├── internal/
│   ├── config/
│   │   └── config.go                  # 配置管理
│   │
│   ├── server/
│   │   ├── gin_server.go              # HTTP服务器
│   │   └── websocket_handler.go       # WebSocket处理
│   │
│   ├── websocket/
│   │   ├── manager.go                 # WebSocket连接管理
│   │   ├── pool.go                    # 租户线程池
│   │   └── message.go                 # 消息类型定义
│   │
│   ├── agent/
│   │   ├── runtime.go                 # Agent运行时
│   │   ├── session.go                 # 会话管理
│   │   ├── memory.go                  # 记忆管理
│   │   ├── tools.go                   # 工具注册表
│   │   └── prompts.go                 # Prompt模板
│   │
│   ├── llm/
│   │   ├── adapter.go                 # LLM适配器接口
│   │   ├── glm_adapter.go             # GLM-4.7实现
│   │   ├── fallback_adapter.go        # 备用模型实现
│   │   └── circuit_breaker.go         # 熔断器
│   │
│   ├── cost/
│   │   ├── optimizer.go               # 成本优化器
│   │   ├── cache.go                   # 相似问题缓存
│   │   └── summary.go                 # 历史摘要
│   │
│   ├── emotion/
│   │   ├── detector.go                # 情绪检测器
│   │   └── rule_based.go              # 规则匹配实现
│   │
│   ├── database/
│   │   ├── db.go                      # 数据库连接
│   │   ├── models.go                  # 数据模型
│   │   ├── player_repo.go             # 玩家仓储
│   │   ├── conversation_repo.go       # 对话仓储
│   │   └── audit_repo.go              # 审计日志仓储
│   │
│   ├── knowledge/
│   │   └── loader.go                  # 知识库加载
│   │
│   └── observability/
│       ├── telemetry.go               # OpenTelemetry
│       └── metrics.go                 # Prometheus指标
│
├── pkg/
│   ├── logging/
│   │   └── logger.go                  # 日志工具
│   └── utils/
│       ├── timeout.go                 # 超时控制
│       └── retry.go                   # 重试机制
│
├── data/
│   ├── knowledge/
│   │   ├── game_faq.json              # 游戏FAQ
│   │   ├── game_rules.json            # 游戏规则
│   │   └── scenario_desc.json         # 场景描述
│   │
│   └── migrations/
│       └── init.sql                   # 数据库初始化脚本
│
├── configs/
│   ├── config.yaml                    # 配置文件
│   └── config.example.yaml
│
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

### 2.2 前端模块 (JavaScript)

```
frontend/
├── index.html                         # 入口页面
├── css/
│   └── styles.css                     # 全局样式
│
├── js/
│   ├── main.js                        # 主入口
│   │
│   ├── scenes/
│   │   ├── BootScene.js               # 启动场景
│   │   ├── WaterTownScene.js          # 水乡主场景
│   │   └── UIOverlay.js               # UI覆盖层
│   │
│   ├── entities/
│   │   ├── NPCGuide.js                # NPC少女导游
│   │   ├── Player.js                  # 玩家角色
│   │   └── Background.js              # 背景元素
│   │
│   ├── ui/
│   │   ├── DialogBox.js               # 对话框
│   │   ├── InputBox.js                # 输入框
│   │   └── Typewriter.js              # 打字机效果
│   │
│   ├── network/
│   │   ├── WebSocketClient.js         # WebSocket客户端
│   │   └── MessageHandler.js          # 消息处理
│   │
│   ├── assets/
│   │   ├── AssetLoader.js             # 资源加载器
│   │   └── sprite_animations.js       # 精灵动画配置
│   │
│   └── utils/
│       └── const.js                   # 常量定义
│
├── assets/
│   ├── images/
│   │   ├── bridge.png                 # 桥梁
│   │   ├── river.png                  # 河流
│   │   ├── stone_road.png             # 青石板路
│   │   ├── npc_guide.png              # NPC少女
│   │   ├── boat.png                   # 乌篷船
│   │   └── dialog_box.png             # 对话框背景
│   │
│   └── fonts/
│       └── game_font.ttf              # 游戏字体
│
├── package.json
├── webpack.config.js (可选)
└── README.md
```

---

## 三、接口定义

### 3.1 WebSocket 消息协议

#### 消息格式
```json
{
  "type": "message_type",
  "requestId": "uuid",
  "tenantId": "tenant_001",
  "timestamp": 1718457600000,
  "payload": { ... }
}
```

#### 客户端→服务器消息类型

**1. 连接确认 (CONNECTION)**
```json
{
  "type": "CONNECTION",
  "requestId": "req_001",
  "tenantId": "tenant_001",
  "payload": {
    "playerId": "player_123",
    "nickname": "玩家小明",
    "deviceId": "device_xyz"
  }
}
```

**2. 玩家消息 (CHAT_MESSAGE)**
```json
{
  "type": "CHAT_MESSAGE",
  "requestId": "req_002",
  "tenantId": "tenant_001",
  "payload": {
    "message": "这个游戏怎么玩？",
    "playerId": "player_123"
  }
}
```

**3. 心跳 (PING)**
```json
{
  "type": "PING",
  "requestId": "req_003",
  "tenantId": "tenant_001",
  "payload": {}
}
```

#### 服务器→客户端消息类型

**1. 欢迎消息 (WELCOME)**
```json
{
  "type": "WELCOME",
  "requestId": "req_001",
  "tenantId": "tenant_001",
  "payload": {
    "guideName": "小荷",
    "message": "欢迎来到江南水乡！我是导游小荷，请问有什么可以帮助你的？",
    "isFirstVisit": true,
    "tips": ["点击输入框与小荷对话", "可以问我关于游戏的问题"]
  }
}
```

**2. NPC回复 (NPC_REPLY)**
```json
{
  "type": "NPC_REPLY",
  "requestId": "req_002",
  "tenantId": "tenant_001",
  "payload": {
    "guideName": "小荷",
    "message": "你可以通过键盘WASD或方向键控制角色移动...",
    "emotion": "happy",
    "actions": ["show_tips", "highlight_controls"]
  }
}
```

**3. 错误消息 (ERROR)**
```json
{
  "type": "ERROR",
  "requestId": "req_002",
  "tenantId": "tenant_001",
  "payload": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "请求过于频繁，请稍后再试"
  }
}
```

**4. 心跳响应 (PONG)**
```json
{
  "type": "PONG",
  "requestId": "req_003",
  "tenantId": "tenant_001",
  "payload": {
    "serverTime": 1718457601000
  }
}
```

### 3.2 REST API 接口

#### 健康检查
```
GET /health
Response: {"status": "ok", "version": "1.0.0"}
```

#### 获取指标
```
GET /metrics
Response: Prometheus metrics text format
```

#### 获取审计日志
```
GET /api/v1/audit?tenantId={tenantId}&startDate={date}&endDate={date}&page={page}
Response:
{
  "total": 100,
  "page": 1,
  "pageSize": 20,
  "logs": [
    {
      "id": "log_001",
      "playerId": "player_123",
      "message": "游戏怎么玩",
      "response": "你可以通过...",
      "timestamp": "2024-06-16T10:00:00Z",
      "emotion": "neutral",
      "cost": 0.001
    }
  ]
}
```

---

## 四、数据模型

### 4.1 PostgreSQL 表结构

**players 表**
```sql
CREATE TABLE players (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(32) NOT NULL,
    nickname VARCHAR(64) NOT NULL,
    device_id VARCHAR(128),
    first_visit_time TIMESTAMP NOT NULL DEFAULT NOW(),
    last_visit_time TIMESTAMP NOT NULL DEFAULT NOW(),
    total_dialogues INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_tenant (tenant_id),
    INDEX idx_device (device_id)
);
```

**conversations 表**
```sql
CREATE TABLE conversations (
    id VARCHAR(64) PRIMARY KEY,
    player_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(32) NOT NULL,
    session_id VARCHAR(64) NOT NULL,
    user_message TEXT NOT NULL,
    ai_message TEXT NOT NULL,
    emotion VARCHAR(16),
    tools_used JSONB,
    llm_model VARCHAR(32),
    llm_tokens INT,
    cost DECIMAL(10,6),
    cache_hit BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_player (player_id),
    INDEX idx_session (session_id),
    INDEX idx_tenant (tenant_id),
    INDEX idx_created (created_at)
);
```

**audit_logs 表**
```sql
CREATE TABLE audit_logs (
    id VARCHAR(64) PRIMARY KEY,
    player_id VARCHAR(64),
    tenant_id VARCHAR(32) NOT NULL,
    action VARCHAR(32) NOT NULL,
    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,
    status VARCHAR(16) NOT NULL,
    latency_ms INT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_tenant (tenant_id),
    INDEX idx_action (action),
    INDEX idx_created (created_at)
);
```

### 4.2 内存数据结构

**Session 会话结构**
```go
type Session struct {
    ID          string
    PlayerID    string
    TenantID    string
    Nickname    string
    IsFirstVisit bool
    CreatedAt   time.Time
    LastActive  time.Time
    Messages    []Message
    Context     map[string]interface{}
    Embeddings  []Embedding // 历史问题向量
}

type Message struct {
    Role      string  // "user" | "assistant"
    Content   string
    Timestamp time.Time
    Tools     []ToolCall
    Emotion   string
}
```

**缓存条目结构**
```go
type CacheEntry struct {
    Key        string
    Question   string
    Answer     string
    Embedding  []float32
    Similarity float32
    TTL        time.Duration
    CreatedAt  time.Time
}
```

---

## 五、核心模块实现要点

### 5.1 WebSocket Handler
- 支持 `upgrade` 请求
- 实现心跳机制 (30s 间隔)
- 租户隔离的连接池
- 消息路由到不同处理器

### 5.2 Agent Runtime
- **总超时 30s** (可通过配置调整)
- 消息路由: 确定请求类型 (welcome | chat)
- 工具调用: 并行执行超时控制 (5s)
- LLM 调用: 带重试和超时 (10s)
- 摘要压缩: 历史消息超过 10 条时压缩

### 5.3 Memory (记忆)
- 玩家基本信息 (nickname, first_visit)
- 对话历史 (滚动窗口 + 摘要)
- 可选 Redis 持久化

### 5.4 Tool Registry
工具接口定义:
```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
    Timeout() time.Duration
}
```

实现工具:
1. `get_player_info`: 获取玩家信息
2. `get_game_guide`: 获取游戏指南
3. `detect_emotion`: 检测情绪
4. `get_quest_info`: 获取任务信息

### 5.5 熔断器
- 状态: Closed | Open | Half-Open
- 阈值: 5 次失败 / 30s 窗口
- 开启后调用备用模型

### 5.6 成本优化
- **相似问题缓存**: 使用 Embedding 相似度 > 0.95
- **历史摘要**: 超过 10 条历史消息时压缩
- **调用统计**: Prometheus 指标

### 5.7 资源隔离
- 每个 tenant_id 独立的 goroutine pool
- pool 大小: `max(10, min(100, tenant_player_count))`
- 防止大租户抢占资源

---

## 六、Phaser 前端实现要点

### 6.1 场景架构
```
WaterTownScene (主场景)
├── Background (背景层)
│   ├── Sky
│   ├── River
│   ├── Bridge
│   └── StoneRoad
├── Entities (实体层)
│   ├── NPCGuide (少女导游)
│   ├── Player (玩家)
│   └── Boat (乌篷船)
└── UIOverlay (UI层)
    ├── DialogBox (对话框)
    ├── InputBox (输入框)
    └── Typewriter (打字机效果)
```

### 6.2 NPC 少女导游
- 站在桥头位置
- 4 方向动画 (idle, walk)
- 呼吸动画效果
- 对话时朝向玩家

### 6.3 对话系统
- 打字机效果输出文本
- 支持多行消息
- 情绪表情 (happy, confused, angry)
- 显示 NPC 名字

### 6.4 WebSocket 集成
- 自动重连机制
- 消息队列 (断线重连后发送)
- 心跳保活

---

## 七、知识库 JSON 结构

### game_faq.json
```json
{
  "categories": [
    {
      "name": "基础操作",
      "questions": [
        {
          "q": "游戏怎么玩？",
          "a": "你可以通过键盘WASD或方向键控制角色移动...",
          "tags": ["操作", "新手"]
        },
        {
          "q": "怎么赚钱？",
          "a": "完成任务、参与活动、出售物品都可以赚取金币...",
          "tags": ["经济", "任务"]
        }
      ]
    },
    {
      "name": "任务系统",
      "questions": [
        {
          "q": "有什么任务？",
          "a": "目前有主线任务、支线任务和日常任务...",
          "tags": ["任务", "玩法"]
        }
      ]
    }
  ]
}
```

---

## 八、Docker Compose 配置

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: water_town
      POSTGRES_USER: water_town
      POSTGRES_PASSWORD: password123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/data/migrations:/docker-entrypoint-initdb.d

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: water_town
      DB_USER: water_town
      DB_PASS: password123
      LLM_API_KEY: ${GLM_API_KEY}
    depends_on:
      - postgres
    volumes:
      - ./backend/data:/app/data

  frontend:
    build: ./frontend
    ports:
      - "3000:80"
    depends_on:
      - backend

volumes:
  postgres_data:
```

---

## 九、实施计划

### 阶段 1: 项目初始化
- [ ] 创建项目目录结构
- [ ] 初始化 go.mod 和 package.json
- [ ] 配置 Docker Compose
- [ ] 创建 PostgreSQL 初始化脚本

### 阶段 2: 后端核心
- [ ] 实现配置管理
- [ ] 实现数据库连接和模型
- [ ] 实现 WebSocket Handler
- [ ] 实现 Agent Runtime (不含 LLM)
- [ ] 实现 Tool Registry
- [ ] 实现熔断器
- [ ] 实现成本优化器
- [ ] 实现 LLM Adapter (GLM-4.7)

### 阶段 3: 前端核心
- [ ] 创建 Phaser 项目基础
- [ ] 实现水乡场景 (背景)
- [ ] 实现 NPC 导游角色
- [ ] 实现对话框 UI
- [ ] 实现输入框 UI
- [ ] 实现 WebSocket 客户端
- [ ] 实现消息处理和打字机效果

### 阶段 4: 集成测试
- [ ] 端到端 WebSocket 通信测试
- [ ] Agent 对话流程测试
- [ ] 熔断器降级测试
- [ ] 缓存命中测试
- [ ] 租户隔离测试

### 阶段 5: 文档
- [ ] README 运行指南
- [ ] API 文档
- [ ] 部署说明

---

## 十、验证计划

### 启动验证
```bash
# 启动所有服务
docker-compose up -d

# 检查服务健康
curl http://localhost:8080/health
curl http://localhost:3000
```

### 功能验证
1. 打开浏览器访问 http://localhost:3000
2. 验证场景加载正确 (桥、水、路)
3. 验证 NPC 少女出现
4. 发送欢迎消息,验证 NPC 自动问候
5. 输入 "游戏怎么玩",验证回答正确
6. 输入相同问题第二次,验证缓存命中
7. 检查数据库中审计日志记录

### 性能验证
```bash
# 查看 Prometheus 指标
curl http://localhost:8080/metrics

# 查看审计日志
curl "http://localhost:8080/api/v1/audit?tenantId=tenant_001"
```

---

## 关键文件清单

### 后端核心文件
- `cmd/server/main.go` - 程序入口
- `internal/server/websocket_handler.go` - WebSocket处理
- `internal/agent/runtime.go` - Agent运行时
- `internal/agent/session.go` - 会话管理
- `internal/agent/tools.go` - 工具注册表
- `internal/llm/glm_adapter.go` - LLM适配器
- `internal/llm/circuit_breaker.go` - 熔断器
- `internal/cost/optimizer.go` - 成本优化器
- `internal/database/models.go` - 数据模型
- `data/knowledge/game_faq.json` - 知识库

### 前端核心文件
- `js/main.js` - 主入口
- `js/scenes/WaterTownScene.js` - 主场景
- `js/entities/NPCGuide.js` - NPC角色
- `js/ui/DialogBox.js` - 对话框
- `js/network/WebSocketClient.js` - WebSocket客户端

### 配置文件
- `docker-compose.yml` - 容器编排
- `backend/configs/config.yaml` - 后端配置
- `backend/data/migrations/init.sql` - 数据库初始化
- `frontend/package.json` - 前端依赖
- `backend/go.mod` - Go依赖