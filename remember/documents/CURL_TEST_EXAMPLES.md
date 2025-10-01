# CURL 测试示例

本文档提供所有API接口的curl命令测试示例，方便快速测试和验证接口功能。

## 前置条件

1. 确保服务已启动（对应端口）
2. 在config.yaml中配置正确的认证token
3. 替换示例中的token值为实际配置的token

## 通用参数

所有请求都需要包含认证头：
```bash
-H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk=" \
-H "Content-Type: application/json"
```

## 话题摘要服务 (端口: 7006)

### 1. 上传消息
```bash
curl -X POST http://localhost:7006/topic_summary/upload \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk=" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test_session_123",
    "messages": [
      {
        "role": "user",
        "content": "我喜欢打篮球和看电影"
      },
      {
        "role": "assistant", 
        "content": "运动和文化娱乐都是很好的爱好"
      },
      {
        "role": "user",
        "content": "最近有什么好看的电影推荐吗？"
      }
    ]
  }'
```

### 2. 查询活跃话题
```bash
curl -X GET http://localhost:7006/topic_summary/activate/test_session_123 \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

### 3. 搜索话题
```bash
curl -X GET "http://localhost:7006/topic_summary/search/test_session_123?q=篮球" \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

### 4. 删除会话
```bash
curl -X DELETE http://localhost:7006/topic_summary/delete/test_session_123 \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

## 聊天事件服务 (端口: 7007)

### 1. 上传聊天事件
```bash
curl -X POST http://localhost:7007/chat_event/upload \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk=" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "chat_session_456",
    "messages": [
      {
        "role": "user",
        "content": "你好，今天天气怎么样？"
      },
      {
        "role": "assistant",
        "content": "今天天气晴朗，温度适宜"
      }
    ]
  }'
```

### 2. 查询聊天事件
```bash
curl -X GET http://localhost:7007/chat_event/get/chat_session_456 \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

### 3. 删除聊天事件
```bash
curl -X DELETE http://localhost:7007/chat_event/delete/chat_session_456 \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

## 用户画像服务 (端口: 7008)

### 1. 上传用户消息
```bash
curl -X POST http://localhost:7008/user_poritrait/upload \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk=" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user_profile_789",
    "messages": [
      {
        "role": "user",
        "content": "我平时喜欢编程和阅读技术书籍"
      },
      {
        "role": "assistant",
        "content": "技术学习是很棒的爱好"
      },
      {
        "role": "user",
        "content": "最近在学习Go语言和分布式系统"
      }
    ]
  }'
```

### 2. 查询用户画像
```bash
curl -X GET http://localhost:7008/user_poritrait/get/user_profile_789 \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

### 3. 删除用户画像
```bash
curl -X DELETE http://localhost:7008/user_poritrait/delete/user_profile_789 \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

## 会话消息服务 (端口: 7009)

### 1. 上传会话消息
```bash
curl -X POST http://localhost:7009/session_messages/upload \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk=" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "session_msg_101",
    "messages": [
      {
        "role": "user",
        "content": "这个产品有什么功能？"
      },
      {
        "role": "assistant",
        "content": "我们的产品支持多种功能..."
      }
    ],
    "task_id": "optional_task_id_123"
  }'
```

### 2. 查询会话消息
```bash
curl -X GET http://localhost:7009/session_messages/get/session_msg_101 \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

### 3. 删除会话消息
```bash
curl -X DELETE http://localhost:7009/session_messages/delete/session_msg_101 \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
```

## 批量测试脚本

### 话题摘要服务完整测试
```bash
#!/bin/bash

# 设置token
TOKEN="GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
SESSION_ID="test_session_$(date +%s)"

echo "=== 话题摘要服务测试 ==="

# 1. 上传消息
echo "上传消息..."
curl -X POST http://localhost:7006/topic_summary/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"$SESSION_ID\",
    \"messages\": [
      {
        \"role\": \"user\",
        \"content\": \"测试消息内容1\"
      },
      {
        \"role\": \"assistant\", 
        \"content\": \"测试回复内容1\"
      }
    ]
  }"

echo -e "\n\n"

# 2. 查询活跃话题
echo "查询活跃话题..."
curl -X GET http://localhost:7006/topic_summary/activate/$SESSION_ID \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n"

# 3. 搜索话题
echo "搜索话题..."
curl -X GET "http://localhost:7006/topic_summary/search/$SESSION_ID?q=测试" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n"

# 4. 清理测试数据
echo "删除测试数据..."
curl -X DELETE http://localhost:7006/topic_summary/delete/$SESSION_ID \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n=== 测试完成 ==="
```

### 所有服务快速测试
```bash
#!/bin/bash

TOKEN="GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="
TIMESTAMP=$(date +%s)

echo "=== 所有服务快速测试 ==="

# 测试话题摘要服务
echo "测试话题摘要服务..."
curl -X POST http://localhost:7006/topic_summary/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"test_topic_$TIMESTAMP\",
    \"messages\": [{\"role\": \"user\", \"content\": \"test\"}]
  }" -s | jq '.code'

# 测试聊天事件服务  
echo "测试聊天事件服务..."
curl -X POST http://localhost:7007/chat_event/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"test_chat_$TIMESTAMP\", 
    \"messages\": [{\"role\": \"user\", \"content\": \"test\"}]
  }" -s | jq '.code'

# 测试用户画像服务
echo "测试用户画像服务..."
curl -X POST http://localhost:7008/user_poritrait/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"test_user_$TIMESTAMP\",
    \"messages\": [{\"role\": \"user\", \"content\": \"test\"}]
  }" -s | jq '.code'

# 测试会话消息服务
echo "测试会话消息服务..."
curl -X POST http://localhost:7009/session_messages/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"test_session_$TIMESTAMP\",
    \"messages\": [{\"role\": \"user\", \"content\": \"test\"}]
  }" -s | jq '.code'

echo "=== 所有服务测试完成 ==="
```

## 常见问题排查

### 1. 认证失败
```bash
# 检查token配置
curl -X POST http://localhost:7006/topic_summary/upload \
  -H "Authorization: Bearer wrong_token" \
  -H "Content-Type: application/json" \
  -d '{"session_id": "test", "messages": []}'
```

### 2. 参数错误
```bash
# 缺少必要参数
curl -X POST http://localhost:7006/topic_summary/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"messages": []}'
```

### 3. 服务未启动
```bash
# 检查服务端口
netstat -an | grep 7006
lsof -i :7006

# 检查服务日志
tail -f nohup.out
```

## 性能测试

### 并发上传测试
```bash
#!/bin/bash

TOKEN="GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="

for i in {1..10}; do
  curl -X POST http://localhost:7006/topic_summary/upload \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"session_id\": \"load_test_$i\",
      \"messages\": [{\"role\": \"user\", \"content\": \"消息$i\"}]
    }" &
done

wait
echo "并发测试完成"
```

## 使用提示

1. **安装jq工具** 用于格式化JSON响应：
   ```bash
   brew install jq  # macOS
   apt-get install jq  # Ubuntu
   ```

2. **保存测试脚本** 为可执行文件：
   ```bash
   chmod +x test_script.sh
   ./test_script.sh
   ```

3. **监控服务状态** 使用日志输出：
   ```bash
   tail -f nohup.out
   ```

4. **调试模式** 查看详细请求信息：
   ```bash
   curl -v -X POST http://localhost:7006/topic_summary/upload ...
