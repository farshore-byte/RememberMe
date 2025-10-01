package server

const (
	DB_NAME     = "main"                // 数据库名
	QUEUE_NAME  = "remember:main:queue" // 队列名
	SERVER_NAME = "[主服务]"               // 服务名
	//User_query       = "Returns the English JSON  result" // user_query填充位
	MaxRetry         = 3  // 任务执行大重试次数
	Monitor_Interval = 60 // 监控间隔
	Queue_MAXLEN     = 80 // 队列最大长度

	//--------------------------  任务触发频次 例如5轮一总结 -----------------------------
	// 主题归纳即时性比较强，尽量高频次，清理轮次要比所有任务的轮次都要唱
	//MessagesRound =1    //消息入库
	UserRound  = 1  // 用户画像生成
	EventRound = 5  // 关键事件抽取
	TopicRound = 1  // 主题归纳
	ClearRound = 15 // 清理已完成处理的轮次

)
