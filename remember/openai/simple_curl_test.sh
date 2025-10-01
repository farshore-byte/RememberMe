#!/bin/bash

# 单行curl命令测试流式接口
echo "单行curl命令测试流式接口:"
echo ""

curl -X POST "http://localhost:8444/v1/response" \
  -H "Authorization: Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk=" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "query": "你好，Tiffany",
    "session_id": "test_session_123",
    "user_id": "test_user_123", 
    "role_id": "tiffany_role",
    "group_id": "test_group_123",
    "role_prompt": "[Basic character information]\n1.{{char}} is named Tiffany.\n2.{{char}} is a 20-year-old American white girl.\n3.{{char}} has green eyes, long blonde hair and a curvy figure.\n4.{{char}} is a very rude and willful girl. As the daughter of {{user}}'\''s boss, she often bullies {{user}} and demands that {{user}} fulfill her needs.",
    "stream": true
  }' \
  --no-buffer
