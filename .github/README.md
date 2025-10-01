# RememberMe - Memory-Enhanced Conversational System

[中文文档](README_zh.md) | [English Documentation](README.md)

<div align="center">

![RememberMe](https://img.shields.io/badge/RememberMe-AI-blue?style=for-the-badge&logo=ai&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-4.4+-47A248?style=for-the-badge&logo=mongodb&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-6+-DC382D?style=for-the-badge&logo=redis&logoColor=white)

**Intelligent conversations with persistent memory and personalized experiences**

[Features](#features) • [Quick Start](#quick-start) • [Architecture](#architecture) • [API](#api-1) • [Contributing](#contributing-1)

</div>

## 🚀 Overview

RememberMe is a sophisticated memory-enhanced conversational system that intelligently remembers user interactions, preferences, and key events to deliver personalized and contextually-aware conversations.

## Features

### 🧠 Intelligent Memory
- **Persistent Context**: Maintains conversation history across sessions
- **User Profiling**: Automatically builds detailed user profiles and preferences
- **Topic Tracking**: Identifies and summarizes key discussion topics
- **Event Timeline**: Records important conversations and tasks

### 💬 Advanced Conversation
- **Streaming Responses**: Real-time AI response generation
- **Multi-turn Context**: Maintains context across multiple exchanges
- **Role Customization**: Fully customizable AI personalities
- **Variable Substitution**: Dynamic content replacement in conversations

### 🎨 Modern Interface
- **Responsive Design**: Optimized for desktop and mobile
- **Real-time Updates**: Live memory panel synchronization
- **Multi-language**: Built-in Chinese/English support
- **Clean UI**: Modern, intuitive user interface

## Architecture

### Microservices Design
The system is built with a microservices architecture for scalability and maintainability:

| Service | Port | Description |
|---------|------|-------------|
| **Main Service** | 6006 | Core conversation coordination |
| **Session Service** | 9120 | Conversation history management |
| **User Portrait** | 9121 | User profiling and preference analysis |
| **Topic Summary** | 9122 | Topic extraction and summarization |
| **Chat Events** | 9123 | Event recording and task management |
| **OpenAI Service** | 8344 | LLM integration and response generation |
| **Web Frontend** | 8120 | User interface and real-time chat |

### Technology Stack

**Backend:**
- **Go** - High-performance microservices
- **Redis** - Caching and message queues
- **MongoDB** - Persistent data storage
- **BytePlus/OpenAI** - Large Language Model integration

**Frontend:**
- **React 18** - Modern UI framework
- **Vite** - Fast build tooling
- **CSS3** - Modern styling with animations
- **i18n** - Internationalization support

## Quick Start

### Prerequisites

- Go 1.19+
- Node.js 16+
- Redis 6+
- MongoDB 4.4+

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/farshore-byte/memory-remember.git
cd memory-remember
```

2. **Configure environment**
```bash
# Copy configuration template
cp remember/config.yaml.example remember/config.yaml

# Edit configuration
vim remember/config.yaml
```

3. **Configuration**
Edit `remember/config.yaml`:

```yaml
redis:
  host: localhost
  port: 6379
  db: 2
  password: ""       # Redis password (optional)
  ssl: false

mongodb:
  uri: "mongodb://localhost:27017/"
  tls: false
  db: "remember"

llm:
  ServiceProvider: "byteplus"  # or "openai"
  api_key: "YOUR_API_KEY_HERE"
  base_url: "https://ark.ap-southeast.bytepluses.com/api/v3"
  model_id: "YOUR_MODEL_ID_HERE"
  temperature: 0.4
  top_p: 0.8
  max_new_tokens: 256

feishu:
  webhook: "YOUR_FEISHU_WEBHOOK_URL_HERE"

auth:
  token: "YOUR_AUTH_TOKEN_HERE"

server:
  session_messages: 9120
  user_poritrait: 9121
  topic_summary: 9122
  chat_event: 9123
  openai: 8344
  main: 6006
  web: 8120
```

4. **Install dependencies**
```bash
# Backend dependencies
cd remember
go mod download

# Frontend dependencies
cd ../remember-web
npm install
```

5. **Build the project**
```bash
# Build all services
./build_all.sh

# Or build individually
cd remember
go build -o server_main server_main.go
go build -o messages_main messages_main.go
go build -o user_main user_main.go
go build -o topic_main topic_main.go
go build -o event_main event_main.go
go build -o openai_main openai_main.go
```

6. **Start services**
```bash
# Start all services (recommended)
./service.sh start

# Or start individual services
./service.sh start main      # Main service
./service.sh start session   # Session service
./service.sh start user      # User portrait service
./service.sh start topic     # Topic summary service
./service.sh start event     # Chat events service
./service.sh start all       # All services
```

7. **Access the application**
Open your browser and navigate to: `http://localhost:8120`

## API

For complete API documentation including all microservices endpoints, request/response formats, and detailed examples, please refer to:

**[📖 Complete API Documentation](API_DOCUMENTATION.md)**

### Quick Reference

#### Upload Messages
```http
POST /memory/upload
Content-Type: application/json
Authorization: Bearer {token}

{
  "session_id": "string (optional)",
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

#### Query Memory
```http
POST /memory/query
Content-Type: application/json
Authorization: Bearer {token}

{
  "session_id": "string",
  "query": "string (optional)"
}
```

#### Apply Memory
```http
POST /memory/apply
Content-Type: application/json
Authorization: Bearer {token}

{
  "session_id": "string (optional)",
  "user_id": "string",
  "role_id": "string",
  "group_id": "string",
  "role_prompt": "string",
  "query": "string"
}
```

#### Delete Session
```http
DELETE /memory/delete
Content-Type: application/json
Authorization: Bearer {token}

{
  "session_id": "string"
}
```

### Microservices Overview

| Service | Port | Key Endpoints |
|---------|------|---------------|
| **Main Service** | 6006 | `/memory/upload`, `/memory/query`, `/memory/apply` |
| **Session Service** | 9120 | `/session_messages/upload`, `/session_messages/get/{sessionID}` |
| **User Portrait** | 9121 | `/user_poritrait/upload`, `/user_poritrait/get/{sessionID}` |
| **Topic Summary** | 9122 | `/topic_summary/upload`, `/topic_summary/search/{sessionID}` |
| **Chat Events** | 9123 | `/chat_event/upload`, `/chat_event/get/{sessionID}` |
| **OpenAI Service** | 8344 | `/v1/response` (streaming/non-streaming) |

## 🔧 Service Management

### Using the Service Script
```bash
# Start services
./service.sh start [service]

# Stop services
./service.sh stop [service]

# Restart services
./service.sh restart [service]

# Check status
./service.sh status

# Cleanup PID files
./service.sh cleanup
```

### Available Services
- `main` - Main conversation service
- `session` - Session management service
- `user` - User portrait service
- `topic` - Topic summary service
- `event` - Chat events service

## 🏭 Project Structure

```
memory-remember/
├── remember/                 # Backend services
│   ├── server_main.go       # Main service entry
│   ├── messages_main.go     # Session service
│   ├── user_main.go         # User portrait service
│   ├── topic_main.go        # Topic summary service
│   ├── event_main.go        # Chat events service
│   ├── openai_main.go       # OpenAI service
│   ├── config.yaml          # Configuration
│   └── shared/              # Shared code
├── remember-web/            # Frontend application
│   ├── src/
│   │   ├── App.jsx          # Main application component
│   │   ├── components/      # React components
│   │   ├── services/        # API services
│   │   └── i18n.js          # Internationalization
│   └── package.json
└── service.sh               # Service management script
```

## 🐛 Troubleshooting

### Common Issues

1. **Port Conflicts**
```bash
# Check port usage
lsof -i :port_number

# Modify ports in config.yaml
```

2. **Redis Connection Issues**
- Ensure Redis service is running
- Verify Redis configuration

3. **MongoDB Connection Issues**
- Confirm MongoDB service is running
- Check connection string and authentication

4. **API Call Failures**
- Verify LLM API key configuration
- Check network connectivity

### Logs
```bash
# View service logs
tail -f logs/main.log
tail -f logs/session_messages.log
tail -f logs/user_portrait.log
tail -f logs/topic_summary.log
tail -f logs/chat_event.log
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Contact

- Project Homepage: [GitHub Repository](https://github.com/farshore-byte/memory-remember)
- Issue Tracker: [GitHub Issues](https://github.com/farshore-byte/memory-remember/issues)
- Email: contact@farshore.ai

## 📋 Changelog

### v1.0.0 (2025-10-01)
- 🎉 Initial release
- ✨ Complete microservices architecture
- 🎨 Modern user interface
- 🌐 Multi-language support
- 🔧 Service management scripts

---

<div align="center">

**Built with ❤️ by the RememberMe Team**

[![GitHub stars](https://img.shields.io/github/stars/farshore-byte/memory-remember?style=social)](https://github.com/farshore-byte/memory-remember/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/farshore-byte/memory-remember?style=social)](https://github.com/farshore-byte/memory-remember/network/members)
[![GitHub issues](https://img.shields.io/github/issues/farshore-byte/memory-remember)](https://github.com/farshore-byte/memory-remember/issues)

</div>
