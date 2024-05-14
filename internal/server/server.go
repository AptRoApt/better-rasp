package server

import (
	"better-rasp/internal/storage"

	"github.com/gin-gonic/gin"
)

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
	s.router.GET("/", s.index)
	// /api/schedule/{weekNum}/room/{buildingNum}/{num}/
	s.router.GET("/api/schedule/:weekNum/room/:buildingNum/:roomNum", s.roomScheduleHandler)

	s.router.Run()
}
