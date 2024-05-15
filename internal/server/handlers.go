package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "better-rasp",
	})
}

func (s *Server) roomListHandler(c *gin.Context) {
	buildingNum, err := strconv.Atoi(c.Param("buildingNum"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
		return
	}

	rooms := s.storage.GetRoomsByBuildingNum(buildingNum)

	c.JSON(http.StatusOK, rooms)
}

func (s *Server) roomScheduleHandler(c *gin.Context) {
	// получаем их из ссылки вида /api/schedule/{weekNum}/room/:roomId

	roomId, err := strconv.Atoi(c.Param("roomId"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
		return
	}
	weekNum, err := strconv.Atoi(c.Param("weekNum"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
		return
	}
	weekNum += 34
	if weekNum > 52 {
		weekNum = weekNum - 34 - 18
	}
	lessons := s.storage.GetLessonsByRoom(roomId, weekNum)
	// отдаём lessons json'ом
	c.JSON(http.StatusOK, lessons)
}

func (s *Server) groupScheduleHandler(c *gin.Context) {
	// /api/schedule/:weekNum/group/:groupId
	weekNum, err := strconv.Atoi(c.Param("weekNum"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}

	groupId, err := strconv.Atoi(c.Param("groupId"))
	weekNum += 34
	if weekNum > 52 {
		weekNum = weekNum - 34 - 18
	}
	lessons := s.storage.GetLessonsByGroupId(groupId, weekNum)
	c.JSON(http.StatusOK, lessons)
}

func (s *Server) teacherScheduleHandler(c *gin.Context) {
	// /api/schedule/:weekNum/teacher/:teacherId
	weekNum, err := strconv.Atoi(c.Param("weekNum"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}

	teacherId, err := strconv.Atoi(c.Param("teacherId"))
	weekNum += 34
	if weekNum > 52 {
		weekNum = weekNum - 34 - 18
	}
	lessons := s.storage.GetLessonsByTeacherId(teacherId, weekNum)
	c.JSON(http.StatusOK, lessons)
}

func (s *Server) facultyListHandler(c *gin.Context) {
	faculties := s.storage.GetFaculties()
	c.JSON(http.StatusOK, faculties)
}

func (s *Server) courseListHandler(c *gin.Context) {
	facultyId, err := strconv.Atoi(c.Query("facultyId"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}

	courses := s.storage.GetCourses(facultyId)
	c.JSON(http.StatusOK, courses)
}

func (s *Server) educationTypeHandler(c *gin.Context) {
	facultyId, err := strconv.Atoi(c.Query("facultyId"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}

	course, err := strconv.Atoi(c.Query("course"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}

	educationTypes := s.storage.GetEducationTypes(facultyId, course)
	c.JSON(http.StatusOK, educationTypes)
}

func (s *Server) groupListHandler(c *gin.Context) {
	facultyId, err := strconv.Atoi(c.Query("facultyId"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}

	course, err := strconv.Atoi(c.Query("course"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}

	educationTypeId, err := strconv.Atoi(c.Query("educationTypeId"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}
	groups := s.storage.GetGroups(facultyId, course, educationTypeId)
	c.JSON(http.StatusOK, groups)
}

func (s *Server) cathedraListHander(c *gin.Context) {
	cathedras := s.storage.GetCathedras()
	c.JSON(http.StatusOK, cathedras)
}

func (s *Server) teacherListHandler(c *gin.Context) {
	cathedraId, err := strconv.Atoi(c.Query("cathedraId"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
	}

	teachers := s.storage.GetTeachers(cathedraId)
	c.JSON(http.StatusOK, teachers)
}

func (s *Server) searchHandler(c *gin.Context) {
	query := c.Query("q")

	groups, teachers := s.storage.Search(query)

	response := gin.H{
		"groups":   groups,
		"teachers": teachers,
	}
	c.JSON(http.StatusOK, response)
}
