package topic_summary

const (
	DB_NAME          = "topic_summary" // 数据库名
	DB_NAME_2        = "topic_info"
	QUEUE_NAME       = "remember:topic_summary:queue"     // 队列名
	SERVER_NAME      = "[主题归纳]"                           // 服务名
	User_query       = "Returns the English JSON  result" // user_query填充位
	MaxRetry         = 3                                  // 任务执行大重试次数
	Monitor_Interval = 60                                 // 监控间隔
	Queue_MAXLEN     = 80                                 // 队列最大长度
	MAX_TOPIC_COUNT  = 60                                 // 最大话题数量限制

)
