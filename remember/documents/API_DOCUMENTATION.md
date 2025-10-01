# API 详细文档

## 概述

本文档提供AI消息处理系统所有API接口的详细说明，包括请求参数、响应格式和错误代码。

## 认证方式

所有API请求都需要在Header中包含Bearer Token认证：

```http
Authorization: Bearer your-config-token
Content-Type: application/json
```

## 通用响应格式

### 成功响应
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    // 具体业务数据
  }
}
```

### 错误响应
```json
{
  "code": -1,
  "msg": "错误描述信息",
  "data": {}
}
```

## 话题摘要服务 (Topic Summary)

**基础URL**: `http://localhost:7006/topic_summary`

### 1. 上传消息 - POST `/upload`

上传聊天消息用于话题分析。

**请求体**:
```json
{
  "session_id": "string, 必填, 会话ID",
  "messages": [
    {
      "role": "string, 必填, 角色(user/assistant)",
      "content": "string, 必填, 消息内容"
    }
  ]
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "messages uploaded successfully",
  "data": {
    "task_id": "string, 任务ID"
  }
}
```

**错误码**:
- `-1`: 请求体格式错误
- `-1`: session_id或messages为空

### 2. 查询活跃话题 - GET `/activate/{sessionID}`

查询指定会话的活跃话题信息。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "session_id": "string",
    "active_topics": [
      {
        "topic": "string, 话题内容",
        "score": "float64, 话题活跃度分数",
        "last_updated": "int64, 最后更新时间戳"
      }
    ],
    "created_at": "int64, 创建时间",
    "updated_at": "int64, 更新时间"
  }
}
```

**错误码**:
- `-1`: session_id为空
- `-1`: 未找到话题信息

### 3. 搜索话题 - GET `/search/{sessionID}?q={query}`

搜索指定会话中的相关话题。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

**查询参数**:
- `q`: string, 必填, 搜索关键词

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "topic": "string, 话题内容",
      "summary": "string, 话题摘要",
      "relevance": "float64, 相关度分数",
      "message_count": "int, 相关消息数量"
    }
  ]
}
```

**错误码**:
- `-1`: session_id或查询参数为空
- `-1`: 搜索失败

### 4. 删除会话 - DELETE `/delete/{sessionID}`

删除指定会话的所有话题数据。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

**响应**:
```json
{
  "code": 0,
  "msg": "session topics and queue messages deleted successfully",
  "data": {}
}
```

**错误码**:
- `-1`: session_id为空
- `-1`: 删除失败

## 聊天事件服务 (Chat Event)

**基础URL**: `http://localhost:7007/chat_event`

### 1. 上传聊天事件 - POST `/upload`

上传聊天事件数据。

**请求体**:
```json
{
  "session_id": "string, 必填, 会话ID",
  "messages": [
    {
      "role": "string, 必填, 角色(user/assistant)",
      "content": "string, 必填, 消息内容"
    }
  ]
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "messages uploaded successfully",
  "data": {
    "task_id": "string, 任务ID"
  }
}
```

### 2. 查询聊天事件 - GET `/get/{sessionID}`

查询指定会话的聊天事件。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    // 聊天事件数据
  }
}
```

### 3. 删除聊天事件 - DELETE `/delete/{sessionID}`

删除指定会话的聊天事件数据。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

## 用户画像服务 (User Portrait)

**基础URL**: `http://localhost:7008/user_poritrait`

### 1. 上传用户消息 - POST `/upload`

上传用户消息用于画像分析。

**请求体**:
```json
{
  "session_id": "string, 必填, 会话ID",
  "messages": [
    {
      "role": "string, 必填, 角色(user/assistant)",
      "content": "string, 必填, 消息内容"
    }
  ]
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "messages uploaded successfully",
  "data": {
    "task_id": "string, 任务ID"
  }
}
```

### 2. 查询用户画像 - GET `/get/{sessionID}`

查询指定会话的用户画像。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "user_id": "string, 用户ID",
    "interests": ["string", "用户兴趣标签"],
    "behavior_patterns": {
      // 用户行为模式数据
    },
    "preferences": {
      // 用户偏好数据
    }
  }
}
```

### 3. 删除用户画像 - DELETE `/delete/{sessionID}`

删除指定会话的用户画像数据。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

## 会话消息服务 (Session Messages)

**基础URL**: `http://localhost:7009/session_messages`

### 1. 上传会话消息 - POST `/upload`

上传会话消息数据。

**请求体**:
```json
{
  "session_id": "string, 必填, 会话ID",
  "messages": [
    {
      // 消息对象，支持多种格式
    }
  ],
  "task_id": "string, 可选, 任务ID"
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "messages uploaded successfully",
  "data": {
    "message_ids": ["string", "消息ID数组"],
    "count": "int, 消息数量"
  }
}
```

### 2. 查询会话消息 - GET `/get/{sessionID}`

查询指定会话的所有消息。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "messages": [
      {
        "role": "user",
        "content": "用户消息内容"
      },
      {
        "role": "assistant", 
        "content": "助手回复内容"
      }
    ]
  }
}
```

### 3. 删除会话消息 - DELETE `/delete/{sessionID}`

删除指定会话的所有消息。

**路径参数**:
- `sessionID`: string, 必填, 会话ID

## 错误代码说明

| 错误码 | 描述 | 解决方案 |
|--------|------|----------|
| -1 | 通用错误 | 查看具体错误信息 |
| 401 | 认证失败 | 检查Authorization头格式和token值 |
| 400 | 请求参数错误 | 检查请求体和参数格式 |
| 404 | 资源未找到 | 检查资源ID是否正确 |
| 500 | 服务器内部错误 | 查看服务日志 |

## 速率限制

当前版本未实施严格的速率限制，但建议：
- 单个IP每秒不超过10个请求
- 批量操作使用适当间隔

## 版本信息

- API版本: v1
- 最后更新: 2024-01-01
- 兼容性: 向后兼容

## 技术支持

如有问题请联系：
- 邮箱: support@example.com
- 文档: https://github.com/your-repo/docs
