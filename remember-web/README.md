# Remember-Web 前端应用

这是 RememberMe 项目的前端 React 应用，提供现代化的聊天界面和记忆管理功能。

## 📁 目录结构

### 根目录文件
- `package.json` - 项目依赖和脚本配置
- `package-lock.json` - 依赖锁定文件
- `vite.config.js` - Vite 构建工具配置
- `index.html` - 主 HTML 模板文件
- `test-formatting.html` - 测试页面
- `read-config.js` - 配置读取工具

### src/ - 源代码目录

#### 主应用文件
- `main.jsx` - React 应用入口点
- `App.jsx` - 主应用组件
- `App.css` - 主应用样式
- `index.css` - 全局样式
- `i18n.js` - 国际化配置

#### components/ - React 组件
- `ChatContainer.jsx` - 聊天容器组件
- `ChatArea.jsx` - 聊天区域组件
- `ChatInput.jsx` - 聊天输入组件
- `MessageBubble.jsx` - 消息气泡组件
- `Sidebar.jsx` - 侧边栏组件
- `Header.jsx` - 头部组件
- `Footer.jsx` - 底部组件
- `MemoryPanel.jsx` - 记忆面板组件
- `RoleSelector.jsx` - 角色选择器组件

#### services/ - API 服务
- `api.js` - 主 API 服务，包含所有后端接口调用
- `userService.js` - 用户相关服务

#### hooks/ - React Hooks
- `useChat.js` - 聊天功能的自定义 Hook

#### data/ - 静态数据
- `users.json` - 用户数据配置

## 🚀 快速开始

### 安装依赖
```bash
npm install
```

### 开发模式
```bash
npm run dev
```
应用将在 `http://localhost:8120` 启动

### 构建生产版本
```bash
npm run build
```

### 预览生产版本
```bash
npm run preview
```

## 🎨 功能特性

### 聊天界面
- 实时消息显示
- 流式响应支持
- 消息历史记录
- 多轮对话上下文

### 记忆管理
- 实时记忆面板
- 用户画像展示
- 话题摘要查看
- 聊天事件时间线

### 角色系统
- 多角色支持
- 角色切换
- 自定义角色配置

### 国际化
- 中英文支持
- 动态语言切换
- 本地化界面

## 🔧 技术栈

### 前端框架
- **React 18** - 用户界面库
- **Vite** - 快速构建工具
- **CSS3** - 样式和动画

### 状态管理
- React Hooks (useState, useEffect, useContext)
- 自定义 Hooks

### API 集成
- Fetch API
- 错误处理
- 加载状态管理

### 样式方案
- CSS Modules
- 响应式设计
- 移动端优化

## 📱 界面组件说明

### ChatContainer
- 主聊天容器，管理整体布局
- 协调侧边栏和聊天区域的交互

### ChatArea
- 显示聊天消息列表
- 处理消息滚动和自动滚动

### ChatInput
- 用户输入区域
- 支持发送消息和快捷键

### MemoryPanel
- 显示用户记忆信息
- 实时更新记忆状态

### RoleSelector
- 角色选择和切换
- 角色配置管理

## 🔗 API 集成

### 主要接口
- `/memory/upload` - 上传消息
- `/memory/query` - 查询记忆
- `/memory/apply` - 应用记忆
- `/memory/delete` - 删除会话

### 错误处理
- 网络错误处理
- API 错误响应
- 用户友好的错误提示

## 🌐 部署说明

### 开发环境
- 使用 Vite 开发服务器
- 热重载支持
- 开发工具集成

### 生产环境
- 静态文件构建
- CDN 部署支持
- 环境变量配置

## 🔗 相关链接

- [主项目 README](../README.md)
- [后端服务 README](../remember/README.md)
- [API 文档](../API_DOCUMENTATION.md)
