# RememberMe - 记忆增强对话系统

[English Documentation](./README.md) | [中文文档](./README_zh.md)

<div align="center">

![Farshore](https://img.shields.io/badge/Farshore-AI-blue?style=for-the-badge&logo=ai&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-4.4+-47A248?style=for-the-badge&logo=mongodb&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-6+-DC382D?style=for-the-badge&logo=redis&logoColor=white)

**具备持久记忆和个性化体验的智能对话系统**

[功能特性](#核心功能) • [快速开始](#快速开始) • [系统架构](#系统架构) • [API文档](#api文档) • [贡献指南](#贡献指南)

</div>

## 🚀 项目概述

Farshore AI 是一个先进的记忆增强对话系统，能够智能地记住用户交互、偏好和关键事件，提供个性化和上下文感知的对话体验。

## ✨ 核心功能

### 🧠 智能记忆
- **持久上下文**：跨会话维护对话历史
- **用户画像**：自动构建详细的用户画像和偏好
- **话题追踪**：识别并总结关键讨论话题
- **事件时间线**：记录重要对话和任务

### 💬 高级对话
- **流式回复**：实时AI响应生成
- **多轮上下文**：维护多轮对话的上下文
- **角色定制**：完全可定制的AI个性
- **变量替换**：对话中的动态内容替换

### 🎨 现代界面
- **响应式设计**：桌面和移动端优化
- **实时更新**：实时记忆面板同步
- **多语言支持**：内置中英文支持
- **简洁UI**：现代直观的用户界面

## 🏗️ 系统架构

### 微服务设计
系统采用微服务架构，具备可扩展性和可维护性：

| 服务 | 端口 | 描述 |
|------|------|------|
| **主服务** | 6006 | 核心对话协调 |
| **会话服务** | 9120 | 对话历史管理 |
| **用户画像** | 9121 | 用户画像和偏好分析 |
| **话题摘要** | 9122 | 话题提取和总结 |
| **聊天事件** | 9123 | 事件记录和任务管理 |
| **OpenAI服务** | 8344 | LLM集成和响应生成 |
| **Web前端** | 8120 | 用户界面和实时聊天 |

### 技术栈

**后端:**
- **Go** - 高性能微服务
- **Redis** - 缓存和消息队列
- **MongoDB** - 持久化数据存储
- **BytePlus/OpenAI** - 大语言模型集成

**前端:**
- **React 18** - 现代UI框架
- **Vite** - 快速构建工具
- **CSS3** - 现代样式和动画
- **i18n** - 国际化支持

## 🚀 快速开始

### 环境要求

- Go 1.19+
- Node.js 16+
- Redis 6+
- MongoDB 4.4+

### 安装步骤

1. **克隆仓库**
```bash
git clone https://github.com/farshore-byte/RememberMe.git
cd RememberMe
```

2. **配置环境**
```bash
# 复制配置模板
cp remember/config.yaml.example remember/config.yaml

# 编辑配置
vim remember/config.yaml
```

3. **配置说明**
编辑 `remember/config.yaml`：

```yaml
redis:
  host: localhost
  port: 6379
  db: 2
  password: ""       # Redis密码（可选）
  ssl: false

mongodb:
  uri: "mongodb://localhost:27017/"
  tls: false
  db: "remember"

llm:
  ServiceProvider: "byteplus"  # 或 "openai"
  api_key: "YOUR_API_KEY_HERE"
  base_url:  "OPENAI_API_BASE_URL" #"https://ark.ap-southeast.bytepluses.com/api/v3"
  model_id: "YOUR_MODEL_ID_HERE"
  temperature: 0.4
  top_p: 0.8
  max_new_tokens: 256

feishu:
  webhook: "YOUR_FEISHU_WEBHOOK_URL"

auth:
  token: "YOUR_AUTH_TOKEN"

server:
  session_messages: 9120
  user_poritrait: 9121
  topic_summary: 9122
  chat_event: 9123
  openai: 8344
  main: 6006
  web: 8120
```

4. **安装依赖**
```bash
# 后端依赖
cd remember
go mod download

# 前端依赖
cd ../remember-web
npm install
```

5. **构建项目**
```bash
# 构建所有服务
./build_all.sh

# 或分别构建
cd remember
go build -o server_main server_main.go
go build -o messages_main messages_main.go
go build -o user_main user_main.go
go build -o topic_main topic_main.go
go build -o event_main event_main.go
go build -o openai_main openai_main.go
```

6. **启动服务**
```bash
# 启动所有服务（推荐）
./service.sh start

# 或启动单个服务
./service.sh start main      # 主服务
./service.sh start session   # 会话服务
./service.sh start user      # 用户画像服务
./service.sh start topic     # 话题摘要服务
./service.sh start event     # 聊天事件服务
./service.sh start all       # 所有服务
```

7. **访问应用**
打开浏览器访问：`http://localhost:8120`

## 📚 API文档

完整的API文档包含所有微服务接口、请求/响应格式和详细示例，请参考：

**[📖 完整API文档](../API_DOCUMENTATION.md)**

### 快速参考

#### 上传消息
```http
POST /memory/upload
Content-Type: application/json
Authorization: Bearer {token}

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

#### 查询记忆
```http
POST /memory/query
Content-Type: application/json
Authorization: Bearer {token}

{
  "session_id": "string",
  "query": "string (可选)"
}
```

#### 应用记忆
```http
POST /memory/apply
Content-Type: application/json
Authorization: Bearer {token}

{
  "session_id": "string (可选)",
  "user_id": "string",
  "role_id": "string",
  "group_id": "string",
  "role_prompt": "string",
  "query": "string"
}
```

#### 删除会话
```http
DELETE /memory/delete
Content-Type: application/json
Authorization: Bearer {token}

{
  "session_id": "string"
}
```

### 微服务概览

| 服务 | 端口 | 关键接口 |
|------|------|----------|
| **主服务** | 6006 | `/memory/upload`, `/memory/query`, `/memory/apply` |
| **会话服务** | 9120 | `/session_messages/upload`, `/session_messages/get/{sessionID}` |
| **用户画像** | 9121 | `/user_poritrait/upload`, `/user_poritrait/get/{sessionID}` |
| **话题摘要** | 9122 | `/topic_summary/upload`, `/topic_summary/search/{sessionID}` |
| **聊天事件** | 9123 | `/chat_event/upload`, `/chat_event/get/{sessionID}` |
| **OpenAI服务** | 8344 | `/v1/response` (流式/非流式) |

## 🔧 服务管理

### 使用服务脚本
```bash
# 启动服务
./service.sh start [service]

# 停止服务
./service.sh stop [service]

# 重启服务
./service.sh restart [service]

# 查看状态
./service.sh status

# 清理PID文件
./service.sh cleanup
```

### 可用服务
- `main` - 主对话服务
- `session` - 会话管理服务
- `user` - 用户画像服务
- `topic` - 话题摘要服务
- `event` - 聊天事件服务

## 🏭 项目结构

```
memory-remember/
├── remember/                 # 后端服务
│   ├── server_main.go       # 主服务入口
│   ├── messages_main.go     # 会话服务
│   ├── user_main.go         # 用户画像服务
│   ├── topic_main.go        # 话题摘要服务
│   ├── event_main.go        # 聊天事件服务
│   ├── openai_main.go       # OpenAI服务
│   ├── config.yaml          # 配置文件
│   └── shared/              # 共享代码
├── remember-web/            # 前端应用
│   ├── src/
│   │   ├── App.jsx          # 主应用组件
│   │   ├── components/      # React组件
│   │   ├── services/        # API服务
│   │   └── i18n.js          # 国际化配置
│   └── package.json
└── service.sh               # 服务管理脚本
```

## 🐛 故障排除

### 常见问题

1. **端口冲突**
```bash
# 检查端口使用情况
lsof -i :端口号

# 修改config.yaml中的端口配置
```

2. **Redis连接问题**
- 确保Redis服务正在运行
- 验证Redis配置信息

3. **MongoDB连接问题**
- 确认MongoDB服务正在运行
- 检查连接字符串和认证信息

4. **API调用失败**
- 验证LLM API密钥配置
- 检查网络连接

### 日志查看
```bash
# 查看服务日志
tail -f logs/main.log
tail -f logs/session_messages.log
tail -f logs/user_portrait.log
tail -f logs/topic_summary.log
tail -f logs/chat_event.log
```

## 🤝 贡献指南

我们欢迎贡献！请查看[贡献指南](CONTRIBUTING.md)了解详情。

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

- 项目主页: [GitHub仓库](https://github.com/farshore-byte/RememberMe)
- 问题追踪: [GitHub Issues](https://github.com/farshore-byte/RememberMe/issues)
- 邮箱: contact@farshore.ai

## 📋 更新日志

### v1.0.0 (2025-10-01)
- 🎉 初始版本发布
- ✨ 完整的微服务架构
- 🎨 现代化用户界面
- 🌐 多语言支持
- 🔧 服务管理脚本

---

<div align="center">

**由 Farshore AI 团队用心构建 ❤️**

[![GitHub stars](https://img.shields.io/github/stars/farshore-byte/RememberMe?style=social)](https://github.com/farshore-byte/RememberMe/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/farshore-byte/RememberMe?style=social)](https://github.com/farshore-byte/RememberMe/network/members)
[![GitHub issues](https://img.shields.io/github/issues/farshore-byte/RememberMe)](https://github.com/farshore-byte/RememberMe/issues)

</div>
