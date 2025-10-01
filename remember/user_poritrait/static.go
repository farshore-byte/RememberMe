package user_poritrait

const (
	DB_NAME          = "user_poritrait"                   // 数据库名
	QUEUE_NAME       = "remember:user_poritrait:queue"    // 队列名
	SERVER_NAME      = "[用户画像]"                           // 服务名
	User_query       = "Returns the English JSON  result" // user_query填充位
	MaxRetry         = 3                                  // 任务执行大重试次数
	Monitor_Interval = 60                                 // 监控间隔
	Queue_MAXLEN     = 80                                 // 队列最大长度

)
