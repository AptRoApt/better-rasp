package main

import (
	"better-rasp/internal/parser"
	"better-rasp/internal/server"
	"better-rasp/internal/storage"

	"github.com/sirupsen/logrus"
)

func main() {
	var cfg storage.Config = storage.EnvConfig()
	logger := logrus.New()
	storage := storage.New(cfg, logger)
	parser := parser.New(&storage)
	parser.Start()
	server := server.New(&storage)
	server.Start()
}
