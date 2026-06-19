package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type FileLogger struct {
	*logrus.Logger
	fileHook *lumberjack.Logger
}

// NewFileLogger 创建支持文件输出的日志器
func NewFileLogger(cfg Config, fileCfg FileLoggerConfig) (*FileLogger, error) {
	logger := logrus.New()

	// 设置日志格式
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05",
		})
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 添加文件输出
	if fileCfg.Enabled && fileCfg.Path != "" {
		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(fileCfg.Path), 0755); err != nil {
			return nil, err
		}

		// 创建日志文件轮转器
		fileHook := &lumberjack.Logger{
			Filename:   fileCfg.Path,
			MaxSize:    fileCfg.MaxSize,
			MaxBackups: fileCfg.MaxBackups,
			MaxAge:     fileCfg.MaxAge,
			Compress:   fileCfg.Compress,
		}

		// 添加文件hook
		logger.AddHook(&FileHook{
			fileHook: fileHook,
		})

		return &FileLogger{
			Logger:   logger,
			fileHook: fileHook,
		}, nil
	}

	return &FileLogger{
		Logger: logger,
	}, nil
}

// Close 关闭文件hook
func (fl *FileLogger) Close() error {
	if fl.fileHook != nil {
		return fl.fileHook.Close()
	}
	return nil
}

// FileHook 文件hook
type FileHook struct {
	fileHook *lumberjack.Logger
}

func (h *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *FileHook) Fire(entry *logrus.Entry) error {
	// 构建消息，包含 Message 和 Data 中的键值对
	msg := fmt.Sprintf("[%s] [%s] %s",
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Level.String(),
		entry.Message)

	// 添加键值对
	if len(entry.Data) > 0 {
		msg += " | "
		for key, value := range entry.Data {
			msg += fmt.Sprintf("%s=%v ", key, value)
		}
	}
	msg += "\n"

	if _, err := h.fileHook.Write([]byte(msg)); err != nil {
		return err
	}
	return nil
}