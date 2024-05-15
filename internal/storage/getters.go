package storage

import (
	"better-rasp/internal/models"
	"log"
)

func (s *Storage) Search(someString string) ([]models.Group, []models.Teacher) {
	const groupQuery = `select groups.id, groups.name, faculty_id, faculties.name, 
							   course, education_type_id, education_types.name
						FROM groups
						JOIN education_types ON education_type_id = education_types.id
						JOIN faculties ON faculty_id = faculties.id
						WHERE groups.name LIKE $1;`
	var foundGroups []models.Group
	var foundTeachers []models.Teacher
	rows, err := s.pool.Query(groupQuery, someString+"%")
	if err != nil {
		log.Printf("Ошибка при поиске %s", err.Error())
		return foundGroups, foundTeachers
	}
	defer rows.Close()

	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.Id, &group.Name, &group.Faculty.Id, &group.Faculty.Name,
			&group.Course, &group.EducationType.Id, &group.EducationType.Name)
		if err != nil {
			log.Printf("Ошибка при поиске: %s", err.Error())
			return foundGroups, foundTeachers
		}
		foundGroups = append(foundGroups, group)
	}

	// Получаем учителей
	const cathedraQuery = `SELECT id, name FROM m2m_teachers_cathedras
						   JOIN cathedras ON cathedra_id = cathedras.id
						   WHERE teacher_id = $1;`
	const teacherQuery = "SELECT id, name FROM teachers WHERE name like $1;"
	rows, err = s.pool.Query(teacherQuery, someString+"%")
	if err != nil {
		log.Printf("Ошибка при поиске: %s", err.Error())
		return foundGroups, foundTeachers
	}
	defer rows.Close()

	cathedraStmt, err := s.pool.Prepare(cathedraQuery)
	if err != nil {
		log.Printf("Ошибка при поиске: %s", err.Error())
		return foundGroups, foundTeachers
	}
	defer cathedraStmt.Close()

	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.Id, &teacher.Name)
		if err != nil {
			log.Printf("Ошибка при поиске: %s", err.Error())
			return foundGroups, foundTeachers
		}

		cathedraRows, err := cathedraStmt.Query(teacher.Id)
		if err != nil {
			log.Printf("Ошибка при поиске: %s", err.Error())
			return foundGroups, foundTeachers
		}
		defer cathedraRows.Close()

		for cathedraRows.Next() {
			var cathedra models.Cathedra
			err := cathedraRows.Scan(&cathedra.Id, &cathedra.Name)
			if err != nil {
				log.Printf("Ошибка при поиске: %s", err.Error())
				return foundGroups, foundTeachers
			}
			teacher.Cathedras = append(teacher.Cathedras, cathedra)
		}
		foundTeachers = append(foundTeachers, teacher)
	}
	return foundGroups, foundTeachers
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
	const query = "select id, building_num, num from rooms where building_num = $1 ORDER BY num ASC;"
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

func (s *Storage) GetLessonsByTeacherId(teacherId int, weekNum int) []models.Lesson {
	var lessons []models.Lesson
	const lessonQuery = `SELECT lessons.id,  disciplines.id, disciplines.name,
	lesson_types.id, lesson_types.name, date,
	lesson_num, cathedras.id,
	cathedras.name, rooms.id, rooms.building_num, rooms.num, is_commission
	FROM lessons
	JOIN lesson_types ON lesson_type_id = lesson_types.id
	JOIN lesson_time ON lesson_num = lesson_time.num
	JOIN disciplines ON discipline_id = disciplines.id
	JOIN rooms ON room_id = rooms.id
	JOIN cathedras ON cathedra_id = cathedras.id
	JOIN m2m_teachers_lessons ON lessons.id = lesson_id
	WHERE teacher_id = $1
	AND extract('week' from date) = $2 ORDER BY lessons.date ASC, lesson_num ASC;`

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
		log.Printf("Ошибка при получении расписания учителя: %s", err.Error())
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
		log.Printf("Ошибка при получении расписания учителя: %s", err.Error())
		return lessons
	}
	defer teacherStmt.Close()

	lessonRows, err := s.pool.Query(lessonQuery, teacherId, weekNum)
	if err != nil {
		log.Printf("Ошибка при получении расписания учителя: %s", err.Error())
		return lessons
	}
	defer lessonRows.Close()

	for lessonRows.Next() {
		var lesson models.Lesson
		err := lessonRows.Scan(&lesson.Id, &lesson.Discipline.Id,
			&lesson.Discipline.Name, &lesson.LessonType.Id,
			&lesson.LessonType.Name, &lesson.Date,
			&lesson.LessonNum, &lesson.Cathedra.Id,
			&lesson.Cathedra.Name, &lesson.Room.Id, &lesson.Room.BuildingNum, &lesson.Room.Num, &lesson.IsCommission)
		if err != nil {
			log.Printf("Ошибка при получении расписания учителя: %s", err.Error())
			return lessons
		}

		groupRows, err := groupStmt.Query(lesson.Id)
		if err != nil {
			log.Printf("Ошибка при получении расписания учителя: %s", err.Error())
			return lessons
		}
		defer groupRows.Close()

		for groupRows.Next() {
			var group models.Group
			err := groupRows.Scan(&group.Id, &group.Name, &group.Faculty.Id,
				&group.Faculty.Name, &group.Course,
				&group.EducationType.Id, &group.EducationType.Name)
			if err != nil {
				log.Printf("Ошибка при получении расписания учителя(группы): %s", err.Error())
				return lessons
			}
			lesson.Groups = append(lesson.Groups, group)
		}

		teacherRows, err := teacherStmt.Query(lesson.Id)
		if err != nil {
			log.Printf("Ошибка при получении расписания учителя: %s", err.Error())
			return lessons
		}
		defer teacherRows.Close()

		for teacherRows.Next() {
			var teacher models.Teacher
			err := teacherRows.Scan(&teacher.Id, &teacher.Name)
			if err != nil {
				log.Printf("Ошибка при получении расписания учителя: %s", err.Error())
				return lessons
			}
			lesson.Teachers = append(lesson.Teachers, teacher)
		}
		lessons = append(lessons, lesson)
	}
	return lessons
}

func (s *Storage) GetLessonsByGroupId(groupId int, weekNum int) []models.Lesson {
	var lessons []models.Lesson
	const lessonQuery = `SELECT lessons.id,  disciplines.id, disciplines.name,
	lesson_types.id, lesson_types.name, date,
	lesson_num, cathedras.id,
	cathedras.name, rooms.id, rooms.building_num, rooms.num, is_commission
	FROM lessons
	JOIN lesson_types ON lesson_type_id = lesson_types.id
	JOIN lesson_time ON lesson_num = lesson_time.num
	JOIN disciplines ON discipline_id = disciplines.id
	JOIN rooms ON room_id = rooms.id
	JOIN cathedras ON cathedra_id = cathedras.id
	JOIN m2m_groups_lessons ON lessons.id = lesson_id
	WHERE group_id = $1
	AND extract('week' from date) = $2 ORDER BY lessons.date ASC, lesson_num ASC;`

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

	lessonRows, err := s.pool.Query(lessonQuery, groupId, weekNum)
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
			&lesson.Cathedra.Name, &lesson.Room.Id, &lesson.Room.BuildingNum, &lesson.Room.Num, &lesson.IsCommission)
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

func (s *Storage) GetLessonsByRoom(roomId int, weekNum int) []models.Lesson {
	var lessons []models.Lesson
	const lessonQuery = `SELECT lessons.id,  disciplines.id, disciplines.name,
	lesson_types.id, lesson_types.name, date,
	lesson_num, cathedras.id,
	cathedras.name, rooms.id, rooms.building_num, rooms.num, is_commission
	FROM lessons
	JOIN lesson_types ON lesson_type_id = lesson_types.id
	JOIN lesson_time ON lesson_num = lesson_time.num
	JOIN disciplines ON discipline_id = disciplines.id
	JOIN rooms ON room_id = rooms.id
	JOIN cathedras ON cathedra_id = cathedras.id
	WHERE rooms.id = $1
	AND extract('week' from date) = $2 ORDER BY lessons.date ASC, lesson_num ASC;`

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

	lessonRows, err := s.pool.Query(lessonQuery, roomId, weekNum)
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
			&lesson.Cathedra.Name, &lesson.Room.Id, &lesson.Room.BuildingNum, &lesson.Room.Num, &lesson.IsCommission)
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

func (s *Storage) GetCathedras() []models.Cathedra {
	var cathedras []models.Cathedra
	const query = "SELECT id, name FROM cathedras;"
	rows, err := s.pool.Query(query)
	if err != nil {
		log.Printf("Ошибка при получении кафедр: %s", err.Error())
		return cathedras
	}
	defer rows.Close()

	for rows.Next() {
		var cathedra models.Cathedra
		err := rows.Scan(&cathedra.Id, &cathedra.Name)
		if err != nil {
			log.Printf("Ошибка при получении кафедр: %s", err.Error())
			return cathedras
		}
		cathedras = append(cathedras, cathedra)
	}
	return cathedras
}

func (s *Storage) GetTeachers(cathedraId int) []models.Teacher {
	var teachers []models.Teacher
	const teacherQuery = `SELECT teachers.id, teachers.name FROM m2m_teachers_cathedras
						  JOIN teachers ON teacher_id = teachers.id
						  WHERE cathedra_id = $1;`
	rows, err := s.pool.Query(teacherQuery, cathedraId)
	if err != nil {
		log.Printf("Ошибка при получении учителей: %s", err.Error())
		return teachers
	}
	defer rows.Close()

	const cathedraQuery = `SELECT id, name FROM m2m_teachers_cathedras
	JOIN cathedras ON cathedra_id = cathedras.id
	WHERE teacher_id = $1;`
	cathedraStmt, err := s.pool.Prepare(cathedraQuery)
	if err != nil {
		log.Printf("Ошибка при получении учителей: %s", err.Error())
		return teachers
	}
	defer cathedraStmt.Close()

	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.Id, &teacher.Name)
		if err != nil {
			log.Printf("Ошибка при получении учителей: %s", err.Error())
			return teachers
		}

		cathedraRows, err := cathedraStmt.Query(teacher.Id)
		if err != nil {
			log.Printf("Ошибка при получении учителей: %s", err.Error())
			return teachers
		}
		defer cathedraRows.Close()

		for cathedraRows.Next() {
			var cathedra models.Cathedra
			err := cathedraRows.Scan(&cathedra.Id, &cathedra.Name)
			if err != nil {
				log.Printf("Ошибка при получении учителей: %s", err.Error())
				return teachers
			}
			teacher.Cathedras = append(teacher.Cathedras, cathedra)
		}
		teachers = append(teachers, teacher)
	}
	return teachers
}

func (s *Storage) GetFaculties() []models.Faculty {
	var faculties []models.Faculty
	const query = "SELECT id, name FROM faculties;"
	rows, err := s.pool.Query(query)
	if err != nil {
		log.Printf("Ошибка при получении факультетов: %s", err.Error())
		return faculties
	}
	defer rows.Close()

	for rows.Next() {
		var faculty models.Faculty
		err := rows.Scan(&faculty.Id, &faculty.Name)
		if err != nil {
			log.Printf("Ошибка при получении факультетов: %s", err.Error())
			return faculties
		}
		faculties = append(faculties, faculty)
	}
	return faculties
}

func (s *Storage) GetCourses(facultyId int) []int {
	var courses []int
	const query = "SELECT DISTINCT course FROM groups WHERE faculty_id = $1;"
	rows, err := s.pool.Query(query, facultyId)
	if err != nil {
		log.Printf("Ошибка при получении факультетов: %s", err.Error())
		return courses
	}
	defer rows.Close()

	for rows.Next() {
		var course int
		err := rows.Scan(&course)
		if err != nil {
			log.Printf("Ошибка при получении факультетов: %s", err.Error())
			return courses
		}
		courses = append(courses, course)
	}
	return courses
}

func (s *Storage) GetEducationTypes(facultyId int, course int) []models.EducationType {
	var education_types []models.EducationType
	const query = `SELECT DISTINCT education_types.id, education_types.name FROM groups
				   JOIN education_types ON education_type_id = education_types.id
				   WHERE faculty_id=$1 AND course=$2;`
	rows, err := s.pool.Query(query, facultyId, course)
	if err != nil {
		log.Printf("Ошибка при получении видов обучения: %s", err.Error())
		return education_types
	}
	defer rows.Close()

	for rows.Next() {
		var education_type models.EducationType
		err := rows.Scan(&education_type.Id, &education_type.Name)
		if err != nil {
			log.Printf("Ошибка при получении видов обучения: %s", err.Error())
			return education_types
		}
		education_types = append(education_types, education_type)
	}
	return education_types
}

func (s *Storage) GetGroups(facultyId int, course int, education_type_id int) []models.Group {
	var groups []models.Group
	const query = `select groups.id, groups.name, faculty_id, faculties.name, 
					course, education_type_id, education_types.name
					FROM groups
					JOIN education_types ON education_type_id = education_types.id
					JOIN faculties ON faculty_id = faculties.id
					WHERE faculty_id=$1 AND course=$2 AND education_type_id=$3`
	rows, err := s.pool.Query(query, facultyId, course, education_type_id)
	if err != nil {
		log.Printf("Ошибка при получении групп: %s", err.Error())
		return groups
	}
	defer rows.Close()

	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.Id, &group.Name, &group.Faculty.Id, &group.Faculty.Name,
			&group.Course, &group.EducationType.Id, &group.EducationType.Name)
		if err != nil {
			log.Printf("Ошибка при получении групп: %s", err.Error())
			return groups
		}
		groups = append(groups, group)
	}
	return groups
}
