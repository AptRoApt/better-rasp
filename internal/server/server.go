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
	s.router.GET("/api/schedule/:weekNum/room/:roomId", s.roomScheduleHandler)
	s.router.GET("/api/schedule/:weekNum/group/:groupId", s.groupScheduleHandler)
	s.router.GET("/api/schedule/:weekNum/teacher/:teacherId", s.teacherScheduleHandler)
	s.router.GET("/api/rooms/:buildingNum", s.roomListHandler)
	s.router.GET("api/groups/getFaculties", s.facultyListHandler)
	s.router.GET("/api/groups/getCourses", s.courseListHandler)
	s.router.GET("/api/groups/getEducationTypes", s.educationTypeHandler)
	s.router.GET("/api/groups/getGroups", s.groupListHandler)
	s.router.GET("/api/teachers/getCathedras", s.cathedraListHander)
	s.router.GET("/api/teachers/getTeachers", s.teacherListHandler)
	s.router.GET("/api/search", s.searchHandler)
	s.router.Run()
}
