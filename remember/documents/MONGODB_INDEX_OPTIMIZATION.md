# MongoDB 索引优化说明文档

## 概述

本文档总结了项目中所有MongoDB集合的查询操作类型和字段，为性能优化提供索引创建建议。

## 集合查询分析

### 1. chat_event 集合

**集合名称**: `chat_event` (由 `DB_NAME` 配置决定)

**数据结构**:
```go
type ChatEvent struct {
    ID            string    `bson:"_id"`            // 唯一主键
    SessionID     string    `bson:"session_id"`     // 会话 ID
    CreatedAt     time.Time `bson:"created_at"`     // 创建时间
    Event         string    `bson:"event"`          // 事件描述
    ExecutionTime time.Time `bson:"execution_time"` // 事件发生时间
    EventType     int       `bson:"event_type"`     // 1: 已完成事件, 2: 代办事项
}
```

**查询操作分析**:

| 查询类型 | 查询字段 | 排序字段 | 说明 |
|---------|---------|---------|------|
| 聚合查询 | `session_id`, `event_type` | `execution_time` | 按会话ID查询，按事件类型分组，按执行时间倒序 |
| 删除操作 | `session_id` | - | 删除指定会话的所有事件 |

**推荐索引**:
```javascript
// 复合索引 - 支持主要查询
db.chat_event.createIndex({ "session_id": 1, "event_type": 1, "execution_time": -1 })

// 单字段索引 - 支持删除操作
db.chat_event.createIndex({ "session_id": 1 })
```

### 2. session_messages 集合

**集合名称**: `session_messages` (由 `DB_NAME` 配置决定)

**数据结构**:
```go
type MemoryMessage struct {
    ID               string    `bson:"_id,omitempty"`     // MongoDB 唯一主键
    SessionID        string    `bson:"session_id"`        // 会话 ID
    UserContent      string    `bson:"user_content"`      // 用户输入
    AssistantContent string    `bson:"assistant_content"` // 助手回复
    CreatedAt        time.Time `bson:"created_at"`        // 创建时间
    MessagesID       string    `bson:"messages_id"`       // 消息轮次ID
    Task1            string    `bson:"task1_id"`          // 任务1 用户画像
    Task2            string    `bson:"task2_id"`          // 任务2 关键事件
    Task3            string    `bson:"task3_id"`          // 任务3 主题归纳
    Task4            string    `bson:"task4_id"`          // 任务4 预留位
    Status           int       `bson:"status"`            // 状态
}
```

**查询操作分析**:

| 查询类型 | 查询字段 | 排序字段 | 说明 |
|---------|---------|---------|------|
| 查询操作 | `session_id` | `created_at` | 按会话ID查询消息，按创建时间升序 |
| 更新操作 | `_id` | - | 按消息ID更新状态 |
| 删除操作 | `session_id` | - | 删除指定会话的所有消息 |
| 统计操作 | `session_id` | - | 统计指定会话的消息数量 |
| 复杂查询 | `session_id`, `taskN_id` | - | 查询taskN_id为空的消息 |

**推荐索引**:
```javascript
// 主要查询索引
db.session_messages.createIndex({ "session_id": 1, "created_at": 1 })

// 更新操作索引
db.session_messages.createIndex({ "_id": 1 })

// 清理逻辑索引
db.session_messages.createIndex({ 
    "session_id": 1, 
    "task1_id": 1, 
    "task2_id": 1, 
    "task3_id": 1 
})

// 任务标记查询索引
db.session_messages.createIndex({ 
    "session_id": 1, 
    "task1_id": 1 
})
db.session_messages.createIndex({ 
    "session_id": 1, 
    "task2_id": 1 
})
db.session_messages.createIndex({ 
    "session_id": 1, 
    "task3_id": 1 
})
```

### 3. topic_summary 集合

**集合名称**: `topic_summary` (由 `DB_NAME` 配置决定)

**数据结构**:
```go
type TopicRecord struct {
    ID        string    `bson:"_id"`        // 唯一主键
    SessionID string    `bson:"session_id"` // 会话 ID
    Topic     string    `bson:"topic"`      // 话题名称
    Content   string    `bson:"content"`    // 内容
    Keywords  []string  `bson:"keywords"`   // 关键词
    CreatedAt time.Time `bson:"created_at"` // 创建时间
    UpdatedAt time.Time `bson:"updated_at"` // 更新时间
}
```

**查询操作分析**:

| 查询类型 | 查询字段 | 排序字段 | 说明 |
|---------|---------|---------|------|
| 文本搜索 | `session_id`, `$text` | `score` | **已支持倒排索引**，使用MongoDB全文搜索功能 |
| 精确查询 | `session_id`, `topic` | - | 按会话ID和话题名称查询 |
| 删除操作 | `session_id` | - | 删除指定会话的所有话题 |

**当前倒排索引支持情况**:
- ✅ **topic字段**: 已支持全文搜索倒排索引
- ✅ **content字段**: 已支持全文搜索倒排索引  
- ✅ **keywords字段**: 已支持全文搜索倒排索引
- ✅ **权重设置**: topic权重10，content权重5，keywords权重8

**推荐索引**:
```javascript
// 全文搜索索引（倒排索引）- 已实现
db.topic_summary.createIndex({ 
    "topic": "text",
    "content": "text", 
    "keywords": "text"
}, {
    weights: {
        topic: 10,      // 话题名称权重最高
        keywords: 8,    // 关键词权重次高
        content: 5      // 内容权重相对较低
    },
    name: "text_search_idx"
})

// 复合索引 - 支持精确查询
db.topic_summary.createIndex({ "session_id": 1, "topic": 1 })

// 单字段索引 - 支持删除操作
db.topic_summary.createIndex({ "session_id": 1 })
```

**倒排索引查询示例**:
```javascript
// 代码中实际使用的查询方式
db.topic_summary.find({
    "session_id": "session123",
    "$text": { "$search": "关键词1 关键词2" }
}).sort({ "score": { "$meta": "textScore" } })
```

### 4. topic_info 集合

**集合名称**: `topic_info` (由 `DB_NAME_2` 配置决定)

**数据结构**:
```go
type TopicInfo struct {
    SessionID    string        `bson:"session_id"`    // 会话 ID
    TopicCount   int           `bson:"topic_count"`   // 话题总数
    ActiveTopics []ActiveTopic `bson:"active_topics"` // 活跃话题
    UpdatedAt    time.Time     `bson:"updated_at"`    // 更新时间
}

type ActiveTopic struct {
    Topic      string    `bson:"topic"`       // 话题名称
    LastActive time.Time `bson:"last_active"` // 最近活跃时间
}
```

**查询操作分析**:

| 查询类型 | 查询字段 | 排序字段 | 说明 |
|---------|---------|---------|------|
| 查询操作 | `session_id` | - | 按会话ID查询话题信息 |
| 更新操作 | `session_id` | - | 按会话ID更新话题信息 |
| 删除操作 | `session_id` | - | 删除指定会话的话题信息 |

**推荐索引**:
```javascript
// 主要查询索引
db.topic_info.createIndex({ "session_id": 1 })
```

### 5. user_portrait 集合

**集合名称**: `user_portrait` (由 `DB_NAME` 配置决定)

**数据结构**:
```go
type UserPortrait struct {
    ID           string                 `bson:"_id"`           // 用户 ID
    SessionID    string                 `bson:"session_id"`    // 会话 ID
    UserPortrait map[string]interface{} `bson:"user_portrait"` // 用户画像
    CreatedAt    time.Time              `bson:"created_at"`    // 创建时间
    UpdatedAt    time.Time              `bson:"updated_at"`    // 更新时间
}
```

**查询操作分析**:

| 查询类型 | 查询字段 | 排序字段 | 说明 |
|---------|---------|---------|------|
| 查询操作 | `session_id` | - | 按会话ID查询用户画像 |
| 更新操作 | `session_id` | - | 按会话ID更新用户画像 |
| 删除操作 | `session_id` | - | 删除指定会话的用户画像 |

**推荐索引**:
```javascript
// 主要查询索引
db.user_portrait.createIndex({ "session_id": 1 })
```

## 索引创建策略

### 1. 高优先级索引（立即创建）

```javascript
// chat_event 集合
db.chat_event.createIndex({ "session_id": 1, "event_type": 1, "execution_time": -1 })

// session_messages 集合  
db.session_messages.createIndex({ "session_id": 1, "created_at": 1 })

// topic_summary 集合
db.topic_summary.createIndex({ "session_id": 1, "topic": 1 })

// user_portrait 集合
db.user_portrait.createIndex({ "session_id": 1 })

// topic_info 集合
db.topic_info.createIndex({ "session_id": 1 })
```

### 2. 中优先级索引（性能优化）

```javascript
// session_messages 任务相关索引
db.session_messages.createIndex({ 
    "session_id": 1, 
    "task1_id": 1, 
    "task2_id": 1, 
    "task3_id": 1 
})

// topic_summary 全文搜索索引
db.topic_summary.createIndex({ 
    "session_id": 1,
    "topic": "text",
    "content": "text",
    "keywords": "text"
})
```

### 3. 索引创建注意事项

1. **会话ID优先**: 所有查询都基于 `session_id`，因此所有复合索引都应将会话ID作为第一个字段
2. **排序字段**: 对于需要排序的查询，将排序字段放在索引的最后
3. **文本索引**: 全文搜索索引需要单独创建，不能与其他字段组合
4. **索引大小**: 监控索引大小，避免过度索引影响写入性能

## 性能监控建议

1. **使用 explain() 分析查询计划**
2. **监控慢查询日志**
3. **定期检查索引使用情况**
4. **根据数据增长调整索引策略**

## 总结

通过创建上述索引，可以显著提升MongoDB查询性能，特别是针对基于 `session_id` 的查询操作。建议按照优先级顺序逐步创建索引，并在生产环境进行性能测试。
