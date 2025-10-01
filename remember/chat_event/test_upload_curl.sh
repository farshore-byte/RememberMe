#!/bin/bash

# 测试新的 chat_event upload API
# 支持上传一轮完整的对话（多个对话对），每个对话对都有时间戳

# 设置 API 地址和认证 token
API_URL="http://localhost:8080/chat_event/upload"
AUTH_TOKEN="your_auth_token_here"

# 测试数据文件
TEST_DATA_FILE="test_upload_example.json"

# 检查数据文件是否存在
if [ ! -f "$TEST_DATA_FILE" ]; then
    echo "错误: 测试数据文件 $TEST_DATA_FILE 不存在"
    exit 1
fi

echo "测试新的 chat_event upload API"
echo "================================"
echo "API URL: $API_URL"
echo "数据文件: $TEST_DATA_FILE"
echo

# 执行 curl 请求
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d @"$TEST_DATA_FILE"

echo
echo
echo "请求数据格式说明:"
echo "================="
echo "{
  \"session_id\": \"会话ID\",
  \"conversations\": [
    {
      \"timestamp\": 1695700000,  // 对话对的时间戳
      \"messages\": [
        {\"role\": \"user\", \"content\": \"用户消息\"},
        {\"role\": \"assistant\", \"content\": \"助手回复\"}
      ]
    },
    // 可以有多个对话对...
  ]
}"
