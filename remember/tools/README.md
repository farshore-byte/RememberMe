# Tools 工具目录

这个目录包含 RememberMe 项目的各种工具和实用脚本。

## 📁 文件说明

### 数据库工具
- `create_indexes_go.go` - 自动创建 MongoDB 索引
  - 为所有微服务创建必要的数据库索引
  - 优化查询性能
  - 支持文本搜索索引

- `create_text_index_new.go` - 新版文本索引创建工具
  - 改进的索引创建逻辑
  - 更好的错误处理
  - 支持更多索引类型

### 文档
- `README_INDEX_CREATION.md` - 索引创建详细说明
  - 索引创建步骤
  - 性能优化建议
  - 故障排除指南

- `test.md` - 测试文档

## 🛠️ 使用方法

### 创建数据库索引
```bash
# 进入 tools 目录
cd remember/tools

# 运行索引创建工具
go run create_indexes_go.go

# 或者构建后运行
go build -o create_indexes create_indexes_go.go
./create_indexes
```

### 索引创建工具功能
1. **会话消息索引**
   - session_id 索引
   - user_id 索引
   - 时间戳索引

2. **用户画像索引**
   - user_id 索引
   - 画像类型索引

3. **话题摘要索引**
   - session_id 索引
   - 话题关键词索引
   - 文本搜索索引

4. **聊天事件索引**
   - session_id 索引
   - 事件类型索引
   - 时间范围索引

## 🔧 开发说明

### 添加新工具
1. 在 tools 目录下创建新的 Go 文件
2. 实现工具功能
3. 添加相应的文档说明
4. 更新此 README 文件

### 工具设计原则
- **单一职责**: 每个工具专注于一个特定任务
- **易于使用**: 提供清晰的命令行界面
- **错误处理**: 完善的错误处理和日志记录
- **文档完整**: 每个工具都有详细的使用说明

## 📊 性能优化

### 索引策略
- 为常用查询字段创建索引
- 使用复合索引优化复杂查询
- 定期监控索引使用情况
- 删除不必要的索引

### 维护建议
- 定期运行索引创建工具
- 监控数据库性能
- 根据查询模式调整索引策略

## 🔗 相关链接

- [主项目 README](../../README.md)
- [后端架构 README](../README.md)
- [API 文档](../../API_DOCUMENTATION.md)
