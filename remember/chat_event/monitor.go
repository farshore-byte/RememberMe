package chat_event

import (
	"fmt"
	"log"
	"time"
)

// QueueMonitor 监控队列长度并报警
type QueueMonitor struct {
	Queue    *QueueClient
	MaxLen   int64
	Interval time.Duration
	StopCh   chan struct{}
}

// NewQueueMonitor 创建队列监控器
func NewQueueMonitor(interval time.Duration, maxLen int64) *QueueMonitor {
	return &QueueMonitor{
		Queue:    MessageQueue,
		MaxLen:   maxLen,
		Interval: interval,
		StopCh:   make(chan struct{}),
	}
}

// Start 启动队列监控
func (m *QueueMonitor) Start() {
	go func() {
		log.Printf("✅ QueueMonitor started, maxLen=%d, interval=%s", m.MaxLen, m.Interval)
		ticker := time.NewTicker(m.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-m.StopCh:
				log.Println("🛑 QueueMonitor stopped")
				return
			case <-ticker.C:
				length, err := m.Queue.Length()
				if err != nil {
					log.Printf("⚠️ QueueMonitor error getting length: %v", err)
					continue
				}
				log.Printf("📊 Current queue length: %d", length)
				if length > m.MaxLen {
					alertText := fmt.Sprintf("🚨 Queue length too long: %d > %d", length, m.MaxLen)
					go SendFeishuMsgAsync(alertText)
					log.Println("⚠️ QueueMonitor alert sent:", alertText)
				}
			}
		}
	}()
}

// Stop 停止队列监控
func (m *QueueMonitor) Stop() {
	close(m.StopCh)
}
