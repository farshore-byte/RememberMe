package session_messages

const (
	DB_NAME = "session_messages" // 数据库名
	//QUEUE_NAME       = "remember:session_messages:queue" // 队列名
	SERVER_NAME = "[会话消息]" // 服务名
	//MaxRetry         = 3                                 // 任务执行大重试次数
	//Monitor_Interval = 60                                // 监控间隔
	//Queue_MAXLEN     = 80                                // 队列最大长度
	PROJECT_MESSAGES_COUNT = 5 // 清理操作时，强制保留的消息数量
)
