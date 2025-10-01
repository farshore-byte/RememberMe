# AI Message Processing System

一个基于Go语言的AI消息处理系统，提供多个服务模块用于处理聊天消息、话题摘要、用户画像和会话消息管理。

## 系统架构

系统包含以下服务模块：

1. **topic_summary** - 话题摘要服务
2. **chat_event** - 聊天事件处理服务  
3. **user_poritrait** - 用户画像服务
4. **session_messages** - 会话消息管理服务

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置设置

编辑 `config.yaml` 文件：

```yaml
redis:
  host: localhost
  port: 6379
  db: 2
  password: ""
  ssl: false

mongodb:
  uri: "mongodb://127.0.0.1:27017"
  tls: false
  db: "ai-msg"

llm:
  ServiceProvider: "byteplus"
  api_key: "your-api-key"
  base_url: "https://ark.ap-southeast.bytepluses.com/api/v3"
  model_id: "your-model-id"
  temperature: 0.4
  top_p: 0.95
  max_new_tokens: 256

feishu:
  webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook"

auth:
  token: "your-bearer-token"
```

### 3. 运行服务

每个服务模块都可以独立运行：

```bash
# 运行话题摘要服务
go run topic_summary/main.go

# 运行动态事件服务  
go run chat_event/main.go

# 运行用户画像服务
go run user_poritrait/main.go

# 运行会话消息服务
go run session_messages/main.go
```

## API 文档

所有API接口都需要Bearer Token认证：

```
Authorization: Bearer your-token-here
```

### 通用响应格式

```json
{
  "code": 0,      // 0: 成功, -1: 失败
  "msg": "success",
  "data": {}      // 响应数据
}
```

### 话题摘要服务 (topic_summary)

**端口**: 7006

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/topic_summary/upload` | 上传消息用于话题分析 |
| GET | `/topic_summary/activate/{sessionID}` | 查询活跃话题 |
| GET | `/topic_summary/search/{sessionID}?q={query}` | 搜索相关话题 |
| DELETE | `/topic_summary/delete/{sessionID}` | 删除会话所有数据 |

**示例请求**:
```bash
curl -X POST http://localhost:7006/topic_summary/upload \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test_session_123",
    "messages": [
      {"role": "user", "content": "我喜欢打篮球"},
      {"role": "assistant", "content": "运动是个好习惯"}
    ]
  }'
```

### 聊天事件服务 (chat_event)

**端口**: 7007

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/chat_event/upload` | 上传聊天事件 |
| GET | `/chat_event/get/{sessionID}` | 查询聊天事件 |
| DELETE | `/chat_event/delete/{sessionID}` | 删除聊天事件 |

### 用户画像服务 (user_poritrait)

**端口**: 7008

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/user_poritrait/upload` | 上传用户消息用于画像分析 |
| GET | `/user_poritrait/get/{sessionID}` | 查询用户画像 |
| DELETE | `/user_poritrait/delete/{sessionID}` | 删除用户画像数据 |

### 会话消息服务 (session_messages)

**端口**: 7009

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/session_messages/upload` | 上传会话消息 |
| GET | `/session_messages/get/{sessionID}` | 查询会话消息 |
| DELETE | `/session_messages/delete/{sessionID}` | 删除会话消息 |

## 安全认证

系统使用Bearer Token进行API认证：

1. 在config.yaml中配置token
2. 所有API请求必须在Header中包含：
   ```
   Authorization: Bearer your-config-token
   ```

## 数据存储

- **Redis**: 用于消息队列和缓存
- **MongoDB**: 用于持久化存储业务数据

## 错误处理

所有API返回统一的错误格式：

```json
{
  "code": -1,
  "msg": "错误描述",
  "data": {}
}
```

常见错误码：
- `-1`: 通用错误
- `401`: 认证失败

## 开发指南

### 添加新服务模块

1. 创建新的服务目录
2. 复制基础文件结构：
   - `config.go` - 配置管理
   - `api.go` - API路由和处理
   - `db.go` - 数据库操作
   - `main.go` - 服务入口

3. 实现具体的业务逻辑

### 配置说明

每个服务模块独立读取config.yaml配置，支持以下配置项：

- Redis连接配置
- MongoDB连接配置  
- LLM服务配置
- 飞书webhook配置
- 认证token配置

## 部署说明

### 环境要求

- Go 1.18+
- Redis 6.0+
- MongoDB 4.4+
- 网络访问LLM服务权限

### 生产部署

1. 编译二进制文件：
```bash
go build -o topic_summary topic_summary/main.go
```

2. 使用systemd或supervisor管理服务进程

3. 配置反向代理（如nginx）进行负载均衡

## 故障排除

### 常见问题

1. **认证失败**: 检查config.yaml中的token配置
2. **数据库连接失败**: 检查Redis和MongoDB服务状态
3. **端口冲突**: 修改各服务的监听端口

### 日志查看

各服务会在控制台输出运行日志，包含：
- 配置加载状态
- 数据库连接状态
- API请求日志
- 错误信息

## 许可证

MIT License
