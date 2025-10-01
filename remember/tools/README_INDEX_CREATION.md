# MongoDB 索引创建脚本使用说明

## 运行方法

### 方法1：使用 Go 版本（推荐，无需安装 mongo shell）

```bash
# 进入索引创建工具目录
cd tools/index_creator

# 下载依赖
go mod tidy

# 运行索引创建脚本
go run main.go

# 或者编译后运行
go build -o create_indexes
./create_indexes
```

### 方法2：使用 mongo shell 直接运行

```bash
# 连接到MongoDB并运行脚本
mongo your_database_name tools/create_mongo_indexes.js

# 或者指定完整的连接字符串
mongo "mongodb://username:password@host:port/database" tools/create_mongo_indexes.js
```

### 方法2：在 mongo shell 中加载运行

```bash
# 先连接到MongoDB
mongo your_database_name

# 然后在mongo shell中加载并运行脚本
load("tools/create_mongo_indexes.js")
```

### 方法3：使用 MongoDB Compass 运行

1. 打开 MongoDB Compass
2. 连接到您的数据库
3. 点击 "MONGOSH" 标签页
4. 复制脚本内容粘贴到命令行中执行

### 方法4：使用命令行参数（推荐）

```bash
# 基本用法
mongo --quiet tools/create_mongo_indexes.js

# 指定数据库和认证
mongo --username your_username --password your_password --authenticationDatabase admin your_database tools/create_mongo_indexes.js

# 指定主机和端口
mongo --host localhost --port 27017 your_database tools/create_mongo_indexes.js
```

## 连接参数配置

在运行脚本前，请根据您的环境修改连接字符串：

```javascript
// 在 create_mongo_indexes.js 文件中修改这行
const db = connect('mongodb://localhost:27017/remember');

// 常见连接字符串示例：
// 本地开发：mongodb://localhost:27017/remember
// 带认证：mongodb://username:password@localhost:27017/remember
// 集群：mongodb://host1:27017,host2:27017,host3:27017/remember?replicaSet=myReplicaSet
```

## 执行步骤

1. **备份数据库**（重要）：
   ```bash
   mongodump --uri="mongodb://localhost:27017/remember" --out=backup/
   ```

2. **测试连接**：
   ```bash
   mongo --eval "db.version()" your_database
   ```

3. **运行索引创建脚本**：
   ```bash
   mongo --quiet tools/create_mongo_indexes.js
   ```

4. **验证索引创建**：
   ```bash
   # 检查某个集合的索引
   mongo --eval "db.chat_event.getIndexes()" your_database
   ```

## 生产环境注意事项

1. **在业务低峰期执行**：索引创建可能会影响数据库性能

2. **使用后台创建**：脚本已设置 `background: true` 选项

3. **监控进度**：可以查看MongoDB日志监控索引创建进度

4. **分阶段执行**：如果数据量很大，可以分批次创建索引

## 故障排除

### 连接失败
- 检查MongoDB服务是否运行：`systemctl status mongod`
- 验证连接字符串是否正确
- 检查防火墙设置

### 权限问题
- 确保用户有创建索引的权限
- 可能需要 `dbAdmin` 或 `userAdmin` 角色

### 索引已存在
- 脚本会跳过已存在的索引
- 如果需要重新创建，先删除旧索引：
  ```javascript
  db.collection.dropIndex("index_name")
  ```

## 验证索引效果

创建索引后，可以使用 `explain()` 方法验证查询性能：

```javascript
// 在mongo shell中测试查询性能
db.chat_event.find({"session_id": "test_session"}).explain("executionStats")

// 检查是否使用了索引
db.chat_event.find({"session_id": "test_session"}).explain().queryPlanner.winningPlan.inputStage.stage
```

## 常用命令参考

```bash
# 查看所有数据库
mongo --eval "show dbs"

# 查看当前数据库的集合
mongo --eval "show collections" your_database

# 查看集合统计信息
mongo --eval "db.chat_event.stats()" your_database

# 删除索引
mongo --eval "db.chat_event.dropIndex('session_id_idx')" your_database
```

通过以上步骤，您可以顺利运行索引创建脚本并优化MongoDB查询性能。
