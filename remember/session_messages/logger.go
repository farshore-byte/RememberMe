package session_messages

import (
	"log"
)

// ✅ info
func Info(msg string, v ...interface{}) {
	log.Printf("✅ INFO: "+msg, v...)
}

// ⚠️ warning
func Warn(msg string, v ...interface{}) {
	log.Printf("⚠️ WARN: "+msg, v...)
}

// ❌ error
func Error(msg string, v ...interface{}) {
	log.Printf("❌ ERROR: "+msg, v...)
}

// 😅 debug
func Debug(msg string, v ...interface{}) {
	log.Printf("⌛️ DEBUG: "+msg, v...)
}
