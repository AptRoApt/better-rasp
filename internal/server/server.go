package server

import (
	"better-rasp/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

func index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "better-rasp",
	})
}

type Server struct {
	router  *gin.Engine
	storage *storage.Storage
}

func New(s *storage.Storage) Server {
	return Server{
		router:  gin.Default(),
		storage: s,
	}
}

func (s *Server) Start() {
	s.router.LoadHTMLFiles("index.html")
	s.router.Static("/static", "static/")
	s.router.GET("/", index)

	s.router.Run()
}
