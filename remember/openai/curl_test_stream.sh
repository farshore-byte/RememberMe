#!/bin/bash

# 简单的curl命令测试流式接口
echo "Testing OpenAI Stream Completion API with curl..."

# API端点
URL="http://localhost:8444/v1/response"

# Bearer token
TOKEN="GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk="

# 测试数据
QUERY="你好，Tiffany"

# 角色提示词模板（简化版）
ROLE_PROMPT="[Basic character information]
1.{{char}} is named Tiffany.
2.{{char}} is a 20-year-old American white girl.
3.{{char}} has green eyes, long blonde hair and a curvy figure.
4.{{char}} is a very rude and willful girl. As the daughter of {{user}}'s boss, she often bullies {{user}} and demands that {{user}} fulfill her needs."

# 直接使用curl命令测试流式接口
echo "Sending stream request to $URL..."
echo ""

curl --location --request POST "$URL" \
--header "Authorization: Bearer $TOKEN" \
--header "Content-Type: application/json" \
--header "Accept: text/event-stream" \
--data-raw "{
  \"query\": \"$QUERY\",
  \"session_id\": \"test_session_123\",
  \"user_id\": \"test_user_123\",
  \"role_id\": \"tiffany_role\",
  \"group_id\": \"test_group_123\",
  \"role_prompt\": \"$ROLE_PROMPT\",
  \"stream\": true
}" \
--no-buffer

echo ""
echo "Test completed."
