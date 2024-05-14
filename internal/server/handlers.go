package server

import (
	"context"
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

// func findRooms(c *gin.Context) {

// }

func (s *Server) roomScheduleHandler(c *gin.Context) {
	// получаем их из ссылки вида /api/schedule/{weekNum}/room/{buildingNum}/{num}/

	buildingNum, err := strconv.Atoi(c.Param("buildingNum"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
		return
	}
	roomNum := c.Param("roomNum")
	if roomNum == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
		return
	}
	weekNum, err := strconv.Atoi(c.Param("weekNum"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Некорректные параметры запроса"))
		return
	}
	room := s.storage.SaveAndGetRoom(context.TODO(), buildingNum, roomNum)
	if room.Id == 0 {
		c.AbortWithError(http.StatusNotFound, errors.New("Аудитория не найдена"))
		return
	}
	weekNum += 34
	if weekNum > 52 {
		weekNum = weekNum - 34 - 18
	}
	lessons := s.storage.GetLessonsByRoom(room, weekNum)
	// отдаём lessons json'ом
	c.JSON(http.StatusOK, lessons)
}
