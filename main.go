package main

import (
	"better-rasp/internal/parser"
	"better-rasp/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func test(c *gin.Context) {
	c.String(http.StatusOK, "Hello world")
}

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
	router := gin.Default()
	router.GET("/test", test)
	router.Run()
}
