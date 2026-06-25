package loginlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger 登录日志记录器，将登录事件写入独立日志文件
type Logger struct {
	logger *log.Logger
	file   *os.File
	mu     sync.Mutex
}

// New creates a login log recorder. When logFile is empty the logger writes
// only to stdout, which is the desired behaviour for container deployments
// that aggregate logs via the container runtime. When logFile is non-empty the
// logger writes to both the file and stdout so local development keeps seeing
// events in the terminal.
func New(logFile string) (*Logger, error) {
	if logFile == "" {
		l := log.New(os.Stdout, "", 0)
		return &Logger{logger: l}, nil
	}

	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开登录日志文件失败: %w", err)
	}

	// 同时写入文件和标准输出，便于开发时观察
	multi := io.MultiWriter(f, os.Stdout)
	l := log.New(multi, "", 0)

	return &Logger{logger: l, file: f}, nil
}

// Close 关闭日志文件
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
	}
}

// LogEntry 登录日志条目
type LogEntry struct {
	Time      time.Time
	ClientIP  string
	Username  string
	AuthMode  string // local, cas, ldap
	Success   bool
	Reason    string // 失败原因，成功时为空
	UserAgent string
}

// Log 记录一条登录日志
func (l *Logger) Log(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	status := "SUCCESS"
	if !entry.Success {
		status = "FAILED"
	}

	t := entry.Time.Format("2006-01-02 15:04:05")
	l.logger.Printf("[%s] %s | user=%s | mode=%s | ip=%s | ua=%s | reason=%s",
		t, status,
		entry.Username,
		entry.AuthMode,
		entry.ClientIP,
		entry.UserAgent,
		entry.Reason,
	)
}

// LogSuccess 记录登录成功
func (l *Logger) LogSuccess(clientIP, username, authMode, userAgent string) {
	l.Log(LogEntry{
		Time:      time.Now(),
		ClientIP:  clientIP,
		Username:  username,
		AuthMode:  authMode,
		Success:   true,
		UserAgent: userAgent,
	})
}

// LogFailure 记录登录失败
func (l *Logger) LogFailure(clientIP, username, authMode, reason, userAgent string) {
	l.Log(LogEntry{
		Time:      time.Now(),
		ClientIP:  clientIP,
		Username:  username,
		AuthMode:  authMode,
		Success:   false,
		Reason:    reason,
		UserAgent: userAgent,
	})
}
