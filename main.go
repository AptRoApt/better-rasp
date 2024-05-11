package main

import (
	"better-rasp/internal/parser"
	"better-rasp/internal/storage"

	"github.com/sirupsen/logrus"
)

func main() {
	//Подключение storage
	//Запуск и настройка (время запуска) парсеров
	//Запуск сервера
	cfg := storage.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "rasp",
	}
	logger := logrus.New()
	storage := storage.New(cfg, logger)
	parser := parser.New(&storage)
	parser.Start()
}

