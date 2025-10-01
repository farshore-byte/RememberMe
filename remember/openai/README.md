# OpenAI 流式完成接口

基于server服务的apply_memory实现，结合openai.go调用模型方式，创建的类似OpenAI流式完成的接口。

## 功能特性

- ✅ 调用server服务的apply_memory接口获取系统提示词和消息历史
- ✅ 调用OpenAI接口生成回复
- ✅ 流式返回响应内容
- ✅ 自动调用server上传接口保存对话记录
- ✅ 支持角色扮演模板集成

## 接口说明

### 请求端点

```
POST /v1/response
```

### 请求头

```
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json
Accept: text/event-stream
```

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| query | string | 是 | 用户查询内容 |
| session_id | string | 否 | 会话ID，为空时自动生成 |
| user_id | string | 否 | 用户ID |
| role_id | string | 否 | 角色ID |
| group_id | string | 否 | 群组ID |
| role_prompt | string | 否 | 角色提示词模板 |

### 响应格式

流式响应，使用Server-Sent Events (SSE)格式：

```
data: {"code":0,"msg":"success","data":{"content":"回复内容"}}

data: [DONE]
```

## 快速开始

### 1. 启动服务

```bash
cd openai/cmd
go run main.go
```

服务将启动在 `http://localhost:8443`

### 2. 测试接口

```bash
# 给执行权限
chmod +x test_stream.sh

# 运行测试
./test_stream.sh
```

### 3. 使用curl测试

```bash
curl --location --request POST 'http://localhost:8443/v1/response' \
--header 'Authorization: Bearer YOUR_API_KEY' \
--header 'Content-Type: application/json' \
--header 'Accept: text/event-stream' \
--data-raw '{
    "query": "你好，Tiffany",
    "session_id": "test_session_123",
    "user_id": "test_user_123", 
    "role_id": "tiffany_role",
    "group_id": "test_group_123",
    "role_prompt": "[Basic character information]..."
}' \
--no-buffer
```

## 工作流程

1. **获取系统提示词**: 调用server的`/memory/apply`接口，获取角色扮演的系统提示词和历史消息
2. **生成回复**: 使用OpenAI接口生成角色回复
3. **流式返回**: 以SSE格式流式返回生成的回复
4. **保存对话**: 异步调用server的`/memory/upload`接口保存完整的对话记录

## 配置说明

服务使用项目根目录的`config.yaml`配置文件，主要配置项：

```yaml
llm:
  api_key: "your-api-key"
  base_url: "https://api.example.com"
  model_id: "your-model-id"

auth:
  token: "your-bearer-token"

server:
  main: 9100  # server主服务端口
```

## 错误处理

- 认证失败: 返回401状态码
- 参数错误: 返回400状态码和错误信息
- 服务错误: 返回500状态码和错误信息

## 注意事项

- 确保server服务正在运行（端口9100）
- 确保OpenAI API密钥和端点配置正确
- 流式响应需要使用SSE客户端处理
- 对话保存是异步操作，不影响响应速度
