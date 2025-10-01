# Session Messages API

基于 Go 和 MongoDB 的会话消息管理服务，提供消息上传、查询和删除功能。

## API 端点

### 1. 上传消息
上传用户和助手之间的对话消息到数据库。支持 messages 数组格式，自动过滤非支持字段并确保消息成对。

**Endpoint:** `POST /upload`

**请求体:**
```json
{
  "session_id": "string",         // 会话 ID (必需)
  "messages": [                   // 消息数组 (必需)
    {"role": "user", "content": "用户消息内容"},
    {"role": "assistant", "content": "助手回复内容"},
    {"role": "user", "content": "另一个用户消息"}
  ],
  "task_id": "string"             // 任务 ID (可选，可不传)
}
```

**Curl 示例:**
```bash
curl -X POST http://localhost:7006/upload \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session-123",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"},
      {"role": "assistant", "content": "I'\''m doing well, thank you!"},
      {"role": "user", "content": "What can you do?"}
    ],
    "task_id": "task-001"
  }'
```

**特性:**
- 自动过滤非 `role` 和 `content` 字段
- 支持的角色: `user`, `assistant` (其他角色会被忽略)
- 自动处理消息配对，确保用户消息和助手消息成对
- 未配对的用户消息会保存为助手内容为空的消息

**响应示例:**
```json
{
  "code": 0,
  "msg": "message uploaded [会话消息] successfully",
  "data": {
    "message_id": "afae128d-a679-4f37-a297-9845bfb93141"
  }
}
```

### 2. 查询消息
根据 session_id 查询所有相关消息，按创建时间升序排列。

**Endpoint:** `GET /get/{sessionID}`

**Curl 示例:**
```bash
curl -X GET http://localhost:7006/get/test-session-123
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {"content": "Hello, how are you?", "role": "user"},
    {"content": "I'm doing well, thank you!", "role": "assistant"},
    {"content": "What can you do?", "role": "user"},
    {"content": "I can help with various tasks", "role": "assistant"}
  ]
}
```

### 3. 删除消息
删除指定 session_id 的所有消息记录。

**Endpoint:** `DELETE /delete/{sessionID}`

**Curl 示例:**
```bash
curl -X DELETE http://localhost:7006/delete/test-session-123
```

**响应示例:**
```json
{
  "code": 0,
  "msg": "messages deleted successfully",
  "data": {}
}
```

## 错误响应

所有 API 端点使用统一的错误响应格式：

```json
{
  "code": -1,
  "msg": "错误描述",
  "data": {}
}
```

常见错误：
- `session_id is required` - 缺少会话 ID
- `session_id and content are required` - 缺少会话 ID 或内容
- `invalid request body` - 请求体格式错误
- `failed to insert message` - 数据库插入失败
- `failed to get messages` - 数据库查询失败
- `failed to delete messages` - 数据库删除失败

## 运行方式

1. 确保配置文件 `config.yaml` 存在并正确配置 MongoDB 和 Redis
2. 运行主程序：
```bash
go run messages_main.go
```

3. API 服务将在 `http://localhost:7005` 启动

## 数据结构

### MemoryMessage 模型
```go
type MemoryMessage struct {
    ID               string    `bson:"_id,omitempty"`     // MongoDB 唯一主键
    SessionID        string    `bson:"session_id"`        // 会话 ID
    UserContent      string    `bson:"user_content"`      // 用户输入
    AssistantContent string    `bson:"assistant_content"` // 助手回复
    CreatedAt        time.Time `bson:"created_at"`        // 创建时间
    TaskID           string    `bson:"task_id"`           // 任务 ID
    Status           int       `bson:"status"`            // 状态 (1: 已完成, 0: 待处理, -1: 失败)
}
```

## 配置要求

在 `config.yaml` 中需要配置以下内容：

```yaml
redis:
  host: localhost
  port: 6379
  db: 0
  password: ""
  ssl: false

mongodb:
  uri: "mongodb://localhost:27017"
  tls: false
  db: "session_messages"
