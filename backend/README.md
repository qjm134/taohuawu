# 江南水乡智能导游系统 - 后端

基于 Go + Gin + WebSocket + GLM-4.7 的智能导游后端服务。

## 技术栈

- **Go 1.22** - 编程语言
- **Gin** - HTTP 框架
- **Gorilla WebSocket** - WebSocket 支持
- **MySQL** - 数据库
- **GLM-4.7** - 大语言模型
- **Prometheus** - 指标监控

## 项目结构

```
backend/
├── cmd/server/          # 程序入口
├── internal/
│   ├── config/          # 配置管理
│   ├── server/          # HTTP 服务器
│   ├── websocket/       # WebSocket 处理
│   ├── agent/           # Agent 运行时
│   ├── llm/             # LLM 适配器
│   ├── cost/            # 成本优化
│   ├── emotion/         # 情绪检测
│   ├── database/        # 数据库层
│   ├── knowledge/       # 知识库
│   └── observability/   # 可观测性
├── pkg/                 # 工具包
├── data/
│   ├── knowledge/       # 知识库文件
│   └── migrations/      # 数据库迁移
└── configs/             # 配置文件
```

## 快速开始

### 环境要求

- Go 1.22+
- MySQL 8.0+
- GLM-4.7 API Key

备注：
   测试时需要在数据库里创建数据库，表会自动创建
   线上项目已经配置了 Docker Compose，会自动创建数据库

### 配置

1. 复制环境变量示例：
```bash
cp .env.example .env
```

2. 编辑 `.env` 文件，设置你的 GLM API Key：
```
GLM_API_KEY=your_api_key_here
```

### 运行

1. 安装依赖：
```bash
go mod download
```

2. 运行服务：
```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

### Docker 运行

```bash
cd backend
docker build -t watertown-backend .
docker run -p 8080:8080 -e GLM_API_KEY=your_key watertown-backend
```

## API 文档

### WebSocket 连接

**URL:** `ws://localhost:8080/ws/game`

**消息格式:**
```json
{
  "type": "MESSAGE_TYPE",
  "requestId": "req_001",
  "tenantId": "tenant_001",
  "timestamp": 1718457600000,
  "payload": { ... }
}
```

### REST API

**健康检查**
```
GET /health
```

**获取指标**
```
GET /metrics
```

**获取审计日志**
```
GET /api/v1/audit?tenantId={tenantId}&page={page}
```

## 核心特性

### 1. 智能对话
- 基于 GLM-4.7 的自然语言理解
- 上下文记忆和会话管理
- 情绪感知和语气调整

### 2. 成本优化
- 相似问题缓存
- 历史消息摘要
- Token 使用统计

### 3. 高可用
- 熔断器机制
- 备用模型降级
- 自动重试

### 4. 多租户
- 租户隔离
- 独立资源池
- 审计日志

## 开发

### 添加新工具

在 `internal/agent/tools.go` 中添加新的工具：

```go
type MyTool struct{}

func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "工具描述"
}

func (t *MyTool) Timeout() time.Duration {
    return 5 * time.Second
}

func (t *MyTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // 实现工具逻辑
    return result, nil
}

// 在 NewToolRegistry 中注册
registry.Register(&MyTool{})
```

## 许可证

MIT