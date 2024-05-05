package storage

import "better-rasp/internal/models"

func (s *Storage) GetLessonsByGroup(groupId int, weekNum int) ([]models.Lesson, error) {
	panic("Not implemented!")
}

func (s *Storage) GetLessonsByTeacher(teacherId int, weekNum int) ([]models.Lesson, error) {
	panic("Not implemented!")
}

func (s *Storage) GetLessonsByRoom(roomNum int, buildingNum int, weekNum int) ([]models.Lesson, error) {
	panic("Not implemented!")
}
