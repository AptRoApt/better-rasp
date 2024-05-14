package storage

import (
	"better-rasp/internal/models"
	"log"
)

func (s *Storage) RoomsSearch(someNum int) []models.Room {
	const query = "select id, building_num, num from rooms where num = $1 or building_num = $1;"
	var found []models.Room
	rows, err := s.pool.Query(query, someNum)
	if err != nil {
		log.Printf("Ошибка при поиске по копрусам и аудиториям: %s", err.Error())
		return found
	}
	defer rows.Close()

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.Id, &room.BuildingNum, &room.Num)
		if err != nil {
			log.Printf("Ошибка при поиске по копрусам и аудиториям: %s", err.Error())
			return found
		}
		found = append(found, room)
	}
	return found
}

func (s *Storage) GetBuildingNums() []int {
	const query = "SELECT DISTINCT building_num FROM rooms ORDER BY building_num ASC;"
	var buildingNums = make([]int, 0, 8)
	rows, err := s.pool.Query(query)
	if err != nil {
		log.Printf("Ошибка при получении аудиторий: %s", err.Error())
		return buildingNums
	}
	defer rows.Close()

	for rows.Next() {
		var buildingNum int
		err := rows.Scan(&buildingNum)
		if err != nil {
			log.Printf("Ошибка при получении аудиторий: %s", err.Error())
			return buildingNums
		}
		buildingNums = append(buildingNums, buildingNum)
	}
	return buildingNums
}

func (s *Storage) GetRoomsByBuildingNum(buildingNum int) []models.Room {
	const query = "select id, building_num, num from rooms where building_num = $1;"
	var found []models.Room
	rows, err := s.pool.Query(query, buildingNum)
	if err != nil {
		log.Printf("Ошибка при поиске по корпусам: %s", err.Error())
		return found
	}
	defer rows.Close()

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.Id, &room.BuildingNum, &room.Num)
		if err != nil {
			log.Printf("Ошибка при поиске по корпусам: %s", err.Error())
			return found
		}
		found = append(found, room)
	}
	return found
}

func (s *Storage) GetLessonsByRoom(room models.Room, weekNum int) []models.Lesson {
	var lessons []models.Lesson
	const lessonQuery = `SELECT lessons.id,  disciplines.id, disciplines.name,
	lesson_types.id, lesson_types.name, date,
	lesson_num, cathedras.id,
	cathedras.name, is_commission
	FROM lessons
	JOIN lesson_types ON lesson_type_id = lesson_types.id
	JOIN lesson_time ON lesson_num = lesson_time.num
	JOIN disciplines ON discipline_id = disciplines.id
	JOIN rooms ON room_id = rooms.id
	JOIN cathedras ON cathedra_id = cathedras.id
	WHERE rooms.id = $1
	AND extract('week' from date) = $2;`

	const groupQuery = `SELECT groups.id, groups.name, groups.faculty_id,
							   faculties.name, groups.course,
							   groups.education_type_id, education_types.name
						FROM m2m_groups_lessons
						JOIN groups ON group_id=groups.id
						JOIN faculties ON faculty_id = faculties.id
						JOIN education_types ON education_type_id = education_types.id
						WHERE lesson_id = $1;`

	groupStmt, err := s.pool.Prepare(groupQuery)
	if err != nil {
		log.Printf("Ошибка при получении расписания аудитории: %s", err.Error())
		return lessons
	}
	defer groupStmt.Close()

	// Кафедра нам тут не нужна, т.к. хранится в уроке.
	const teacherQuery = `SELECT teachers.id, teachers.name 
						  FROM lessons 
						  JOIN m2m_teachers_lessons ON lessons.id = lesson_id
						  JOIN teachers ON teacher_id = teachers.id
						  WHERE lesson_id = $1;`

	teacherStmt, err := s.pool.Prepare(teacherQuery)
	if err != nil {
		log.Printf("Ошибка при получении расписания аудитории: %s", err.Error())
		return lessons
	}
	defer teacherStmt.Close()

	lessonRows, err := s.pool.Query(lessonQuery, room.Id, weekNum)
	if err != nil {
		log.Printf("Ошибка при получении расписания аудитории: %s", err.Error())
		return lessons
	}
	defer lessonRows.Close()

	for lessonRows.Next() {
		var lesson models.Lesson
		err := lessonRows.Scan(&lesson.Id, &lesson.Discipline.Id,
			&lesson.Discipline.Name, &lesson.LessonType.Id,
			&lesson.LessonType.Name, &lesson.Date,
			&lesson.LessonNum, &lesson.Cathedra.Id,
			&lesson.Cathedra.Name, &lesson.IsCommission)
		if err != nil {
			log.Printf("Ошибка при получении расписания аудитории: %s", err.Error())
			return lessons
		}

		groupRows, err := groupStmt.Query(lesson.Id)
		if err != nil {
			log.Printf("Ошибка при получении расписания аудитории: %s", err.Error())
			return lessons
		}
		defer groupRows.Close()

		for groupRows.Next() {
			var group models.Group
			err := groupRows.Scan(&group.Id, &group.Name, &group.Faculty.Id,
				&group.Faculty.Name, &group.Course,
				&group.EducationType.Id, &group.EducationType.Name)
			if err != nil {
				log.Printf("Ошибка при получении расписания аудитории (группы): %s", err.Error())
				return lessons
			}
			lesson.Groups = append(lesson.Groups, group)
		}

		teacherRows, err := teacherStmt.Query(lesson.Id)
		if err != nil {
			log.Printf("Ошибка при получении расписания аудитории: %s", err.Error())
			return lessons
		}
		defer teacherRows.Close()

		for teacherRows.Next() {
			var teacher models.Teacher
			err := teacherRows.Scan(&teacher.Id, &teacher.Name)
			if err != nil {
				log.Printf("Ошибка при получении расписания аудитории: %s", err.Error())
				return lessons
			}
			lesson.Teachers = append(lesson.Teachers, teacher)
		}
		lessons = append(lessons, lesson)
	}
	return lessons
}
