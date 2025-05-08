package main

import (
	"better-rasp/internal/parser"
	"better-rasp/internal/server"
	"better-rasp/internal/storage"
	"context"

	"github.com/sirupsen/logrus"
)

func main() {
	var cfg storage.Config = storage.EnvConfig()
	logger := logrus.New()
	logger.Level = logrus.InfoLevel
	storage := storage.New(cfg, logger)
	parser := parser.New(&storage)
	ctx := context.TODO()
	parser.Start(ctx)
	server := server.New(&storage)
	server.Start()
}
