package main

import (
	"better-rasp/internal/parser"
	"better-rasp/internal/server"
	"better-rasp/internal/storage"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	file, err := os.Open("config.json")
	if err != nil {
		panic("Невозможно открыть файл с конфигом")
	}
	defer file.Close()

	var cfg storage.Config = storage.EnvConfig()
	logger := logrus.New()
	storage := storage.New(cfg, logger)
	parser := parser.New(&storage)
	parser.Start()
	server := server.New(&storage)
	server.Start()
}
