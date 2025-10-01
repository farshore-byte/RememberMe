package chat_event

const (
	DB_NAME          = "chat_event"                       // 数据库名
	QUEUE_NAME       = "remember:chat_event:queue"        // 队列名
	SERVER_NAME      = "[关键事件]"                           // 服务名
	User_query       = "Returns the English JSON  result" // user_query填充位
	MaxRetry         = 3                                  // 任务执行大重试次数
	Monitor_Interval = 60                                 // 监控间隔
	Queue_MAXLEN     = 80                                 // 队列最大长度

)
