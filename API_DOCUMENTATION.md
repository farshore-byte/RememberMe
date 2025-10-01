# Farshore AI API 文档

## 概述

Farshore AI 是一个基于微服务架构的记忆增强对话系统，包含多个独立的微服务，每个服务负责特定的功能模块。

## 服务架构

| 服务名称 | 端口 | 功能描述 |
|---------|------|----------|
| 主服务 (Main Service) | 6006 | 统一入口，协调各微服务 |
| 会话消息服务 (Session Messages) | 9120 | 管理对话消息存储和查询 |
| 用户画像服务 (User Portrait) | 9121 | 生成和维护用户画像 |
| 话题摘要服务 (Topic Summary) | 9122 | 话题提取和摘要生成 |
| 聊天事件服务 (Chat Event) | 9123 | 管理聊天事件和时间线 |
| OpenAI 服务 (OpenAI Service) | 8344 | 大语言模型接口和流式响应 |
| Web 前端服务 (Web Frontend) | 8120 | 用户界面和代理转发 |

## 认证

所有API都需要Bearer Token认证：

```http
Authorization: Bearer YOUR_AUTH_TOKEN
```

## 主服务 (端口 6006)

### 1. 消息上传接口

**POST** `/memory/upload`

上传对话消息到系统，消息会被分发到各个微服务进行处理。

**请求体：**
```json
{
  "session_id": "string (可选)",
  "user_id": "string",
  "role_id": "string", 
  "group_id": "string",
  "messages": [
    {
      "role": "user|assistant",
      "content": "string"
    }
  ]
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "消息已入队，任务ID：xxx",
  "data": {
    "task_id": "string"
  }
}
```

### 2. 查询接口

**POST** `/memory/query`

获取完整的角色扮演上下文，包括用户画像、话题摘要、聊天事件和会话消息。

**请求体：**
```json
{
  "session_id": "string",
  "query": "string (可选)"
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "user_portrait": {
      // 用户画像数据
    },
    "topic_summary": [
      // 话题摘要列表
    ],
    "chat_events": {
      "todo": [],
      "completed": []
    },
    "session_messages": [
      // 会话消息列表
    ],
    "current_time": "2024-01-01 12:00:00"
  }
}
```

### 3. 获取消息接口

**POST** `/memory/messages`

获取指定session_id或user_id+role_id+group_id的消息。

**请求体：**
```json
{
  "session_id": "string (可选)",
  "user_id": "string (可选)",
  "role_id": "string (可选)", 
  "group_id": "string (可选)"
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "messages": [
      {
        "role": "user|assistant",
        "content": "string",
        "timestamp": "string"
      }
    ]
  }
}
```

### 4. 应用接口

**POST** `/memory/apply`

将记忆应用于系统提示词，并提供历史消息，用于角色扮演场景。

**请求体：**
```json
{
  "session_id": "string (可选)",
  "user_id": "string",
  "role_id": "string",
  "group_id": "string",
  "role_prompt": "string",
  "query": "string"
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "system_prompt": "string",
    "messages": [
      {
        "role": "user|assistant",
        "content": "string"
      }
    ]
  }
}
```

### 5. 删除接口

**DELETE** `/memory/delete`

同时删除所有微服务中的相关数据。

**请求体：**
```json
{
  "session_id": "string"
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "所有微服务数据删除成功",
  "data": {
    "session_id": "string",
    "results": [
      {
        "service_name": "user_portrait",
        "success": true,
        "message": "删除成功"
      }
    ]
  }
}
```

## 会话消息服务 (端口 9120)

### 1. 上传接口

**POST** `/session_messages/upload`

上传会话消息。

**请求体：**
```json
{
  "session_id": "string",
  "messages": [
    {
      "role": "user|assistant",
      "content": "string"
    }
  ],
  "task_id": "string (可选)"
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "messages uploaded successfully",
  "data": {
    "message_ids": ["string"],
    "count": 1
  }
}
```

### 2. 查询接口

**GET** `/session_messages/get/{sessionID}`

获取指定会话的所有消息。

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "messages": [
      {
        "role": "user|assistant",
        "content": "string"
      }
    ]
  }
}
```

### 3. 删除接口

**DELETE** `/session_messages/delete/{sessionID}`

删除指定会话的所有消息。

### 4. 消息计数接口

**GET** `/session_messages/count/{sessionID}`

获取指定会话的消息数量。

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "count": 10
  }
}
```

### 5. 清理接口

**POST** `/session_messages/clean`

清理已处理的消息。

**请求体：**
```json
{
  "session_id": "string"
}
```

### 6. 标记任务接口

**POST** `/session_messages/mark_task`

查找taskN_id为空的消息并标记。

**请求体：**
```json
{
  "session_id": "string",
  "task_index": 1,
  "task_id": "string"
}
```

## 用户画像服务 (端口 9121)

### 1. 上传接口

**POST** `/user_poritrait/upload`

上传消息用于生成用户画像。

**请求体：**
```json
{
  "session_id": "string",
  "messages": [
    {
      "role": "user|assistant",
      "content": "string"
    }
  ]
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "messages uploaded successfully",
  "data": {
    "task_id": "string"
  }
}
```

### 2. 查询接口

**GET** `/user_poritrait/get/{sessionID}`

获取用户画像数据。

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    // 用户画像数据结构
  }
}
```

### 3. 删除接口

**DELETE** `/user_poritrait/delete/{sessionID}`

删除用户画像数据。

## 话题摘要服务 (端口 9122)

### 1. 上传接口

**POST** `/topic_summary/upload`

上传消息用于话题摘要生成。

**请求体：**
```json
{
  "session_id": "string",
  "messages": [
    {
      "role": "user|assistant",
      "content": "string"
    }
  ]
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "messages uploaded successfully",
  "data": {
    "task_id": "string"
  }
}
```

### 2. 活跃话题接口

**GET** `/topic_summary/activate/{sessionID}`

获取活跃话题信息。

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "active_topics": [
      {
        "topic": "string",
        "count": 1
      }
    ]
  }
}
```

### 3. 搜索接口

**GET** `/topic_summary/search/{sessionID}?q={query}`

搜索相关话题。

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "topic": "string",
      "content": "string",
      "timestamp": "string"
    }
  ]
}
```

### 4. 删除接口

**DELETE** `/topic_summary/delete/{sessionID}`

删除会话话题数据。

## 聊天事件服务 (端口 9123)

### 1. 上传接口

**POST** `/chat_event/upload`

上传对话事件。

**请求体：**
```json
{
  "session_id": "string",
  "conversations": [
    {
      "timestamp": 1234567890,
      "messages": [
        {
          "role": "user|assistant",
          "content": "string"
        }
      ]
    }
  ]
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "messages uploaded successfully",
  "data": {
    "task_id": "string"
  }
}
```

### 2. 查询接口

**GET** `/chat_event/get/{sessionID}`

获取聊天事件数据。

**响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "todo": [],
    "completed": []
  }
}
```

### 3. 删除接口

**DELETE** `/chat_event/delete/{sessionID}`

删除聊天事件数据。

## OpenAI 服务 (端口 8344)

### 1. 流式响应接口

**POST** `/v1/response`

获取大语言模型的流式或非流式响应。

**请求体：**
```json
{
  "query": "string",
  "session_id": "string (可选)",
  "user_id": "string (可选)",
  "role_id": "string (可选)",
  "group_id": "string (可选)",
  "role_prompt": "string (可选)",
  "first_message": "string (可选)",
  "stream": true
}
```

**流式响应：**
```
data: {"code":0,"msg":"success","data":{"content":"Hello"}}
data: {"code":0,"msg":"success","data":{"content":" world"}}
data: [DONE]
```

**非流式响应：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "content": "Hello world"
  }
}
```

## Web 前端服务 (端口 8120)

Web前端服务通过代理将请求转发到对应的后端服务：

- `/api/memory/*` → 主服务 (6006)
- `/api/response/*` → OpenAI服务 (8344)

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| -1 | 失败 |

## 使用示例

### 完整对话流程

1. **上传对话消息**
```bash
curl -X POST http://localhost:6006/memory/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "role_id": "role456", 
    "group_id": "group789",
    "messages": [
      {"role": "user", "content": "你好"},
      {"role": "assistant", "content": "你好！很高兴认识你"}
    ]
  }'
```

2. **获取角色扮演上下文**
```bash
curl -X POST http://localhost:6006/memory/query \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "generated_session_id",
    "query": "今天天气怎么样"
  }'
```

3. **获取AI响应**
```bash
curl -X POST http://localhost:8344/v1/response \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "今天天气怎么样",
    "session_id": "generated_session_id",
    "user_id": "user123",
    "role_id": "role456",
    "group_id": "group789",
    "role_prompt": "你是一个友好的助手",
    "stream": false
  }'
```

## 注意事项

1. 所有API都需要Bearer Token认证
2. 端口配置可在 `remember/config.yaml` 中修改
3. 流式响应需要设置 `stream: true` 并处理SSE格式
4. 首次对话可使用 `first_message` 参数设置初始回复
