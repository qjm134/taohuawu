package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/watertown/guide/internal/config"
	"github.com/watertown/guide/internal/database"
	"github.com/watertown/guide/internal/knowledge"
	"github.com/watertown/guide/internal/observability"
	"github.com/watertown/guide/internal/server"
	"github.com/watertown/guide/pkg/logging"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger := logging.New(cfg.Logging)
	logger.Info("Starting Water Town Guide Server...")

	// 初始化数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}

	// 加载知识库
	kb, err := knowledge.Load(cfg.Knowledge.Path)
	if err != nil {
		logger.Fatal("Failed to load knowledge base", "error", err)
	}
	logger.Info("Knowledge base loaded", "questions", len(kb.Categories))

	// 初始化 OpenTelemetry
	tp, err := observability.InitTracing(cfg.Observability)
	if err != nil {
		logger.Warn("Failed to initialize tracing", "error", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = tp.Shutdown(ctx)
		}()
	}

	// 初始化服务器
	srv := server.New(cfg, db, kb, logger)

	// 启动服务器
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("Server failed", "error", err)
		}
	}()

	logger.Info("Server started", "port", cfg.Server.Port)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	}

	logger.Info("Server stopped")
}