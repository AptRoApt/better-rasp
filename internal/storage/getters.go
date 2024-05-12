package storage

import (
	"better-rasp/internal/models"
	"context"
	"log"

	"github.com/lib/pq"
)

const UniqueViolationCode = "23505"

func (s *Storage) GetAllGroups(ctx context.Context) []models.Group {
	const query = `SELECT groups.id, groups.name, faculty_id, f.name, course, education_type_id, et.name FROM groups
				   JOIN faculties f ON faculty_id = f.id
				   JOIN education_types et ON education_type_id = et.id;`
	rows, err := s.pool.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Ошибка при получении всех групп из бд: %s", err.Error())
	}
	defer rows.Close()

	var groups []models.Group

	for rows.Next() {
		var group = models.Group{
			Faculty:       models.Faculty{},
			EducationType: models.EducationType{},
		}
		err := rows.Scan(&group.Id, &group.Name, &group.Faculty.Id, &group.Faculty.Name, &group.Course, &group.EducationType.Id, &group.EducationType.Name)
		if err != nil {
			log.Printf("Ошибка при получении всех групп из бд: %s", err.Error())
		}
		groups = append(groups, group)
	}
	return groups
}

func (s *Storage) GetLessonTypeByName(ctx context.Context, name string) models.LessonType {
	query := "SELECT id, name FROM lesson_types WHERE name=$1;"
	var exists bool
	var lessonType = models.LessonType{
		Name: name,
	}
	rows, err := s.pool.QueryContext(ctx, query, name)
	if err != nil {
		log.Printf("Ошибка при получении типа занятия: %s", err.Error())
		return lessonType
	}
	defer rows.Close()

	for rows.Next() {
		exists = true
		if err := rows.Scan(&lessonType.Id, &lessonType.Name); err != nil {
			log.Printf("Ошибка при получении типа занятия: %s", err.Error())
			return lessonType
		}
	}

	if !exists {
		insQuery := "INSERT INTO lesson_types (name) VALUES ($1) RETURNING id;"
		insRows, err := s.pool.QueryContext(ctx, insQuery, name)
		pqErr, ok := err.(*pq.Error)
		if ok && string(pqErr.Code) == UniqueViolationCode {
			row := s.pool.QueryRowContext(ctx, query, name)
			row.Scan(&lessonType.Id)
		} else if err != nil {
			log.Printf("Ошибка при получении типа занятия: %s", err.Error())
			return lessonType
		} else {
			defer insRows.Close()

			for insRows.Next() {
				if err := insRows.Scan(&lessonType.Id); err != nil {
					log.Printf("Ошибка при получении типа занятия: %s", err.Error())
					return lessonType
				}
			}
		}
		return lessonType
	}

	return lessonType

}

func (s *Storage) GetDisciplineByName(ctx context.Context, name string) models.Discipline {
	query := "SELECT id FROM disciplines WHERE name=$1;"
	var exists bool
	var discipline = models.Discipline{
		Name: name,
	}
	rows, err := s.pool.QueryContext(ctx, query, name)
	if err != nil {
		log.Printf("Ошибка при получении дисциплины: %s", err.Error())
		return discipline
	}
	defer rows.Close()

	for rows.Next() {
		exists = true
		if err := rows.Scan(&discipline.Id); err != nil {
			log.Printf("Ошибка при получении дисциплины: %s", err.Error())
			return discipline
		}
	}

	if !exists {
		insQuery := "INSERT INTO disciplines (name) VALUES ($1) RETURNING id;"
		insRows, err := s.pool.QueryContext(ctx, insQuery, name)
		pqErr, ok := err.(*pq.Error)
		if ok && string(pqErr.Code) == UniqueViolationCode {
			row := s.pool.QueryRowContext(ctx, query, name)
			row.Scan(&discipline.Id)
		} else if err != nil {
			log.Printf("Ошибка при получении дисциплины: %s", err.Error())
			return discipline
		} else {
			defer insRows.Close()
			for insRows.Next() {
				if err := insRows.Scan(&discipline.Id); err != nil {
					log.Printf("Ошибка при получении дисциплины: %s", err.Error())
					return discipline
				}
			}
		}
		return discipline
	}

	return discipline
}

func (s *Storage) SaveAndGetCathedraByName(ctx context.Context, name string) models.Cathedra {
	query := "SELECT id FROM cathedras WHERE name=$1;"
	var exists bool
	var cathedra = models.Cathedra{
		Name: name,
	}
	rows, err := s.pool.QueryContext(ctx, query, name)
	if err != nil {
		log.Printf("Ошибка при получении кафедры: %s", err.Error())
		return cathedra
	}
	defer rows.Close()

	for rows.Next() {
		exists = true
		if err := rows.Scan(&cathedra.Id); err != nil {
			log.Printf("Ошибка при получении кафедры: %s", err.Error())
			return cathedra
		}
	}

	if !exists {
		insQuery := "INSERT INTO cathedras (name) VALUES ($1) RETURNING id;"
		insRows, err := s.pool.QueryContext(ctx, insQuery, name)
		pqErr, ok := err.(*pq.Error)
		if ok && string(pqErr.Code) == UniqueViolationCode {
			row := s.pool.QueryRowContext(ctx, query, name)
			row.Scan(&cathedra.Id)
		} else if err != nil {
			log.Printf("Ошибка при получении кафедры: %s", err.Error())
			return cathedra
		} else {
			defer insRows.Close()

			for insRows.Next() {
				if err := insRows.Scan(&cathedra.Id); err != nil {
					log.Printf("Ошибка при получении кафедры: %s", err.Error())
					return cathedra
				}
			}
		}
		return cathedra
	}

	return cathedra
}

// я ОЧЕНЬ надеюсь, что у нас нет ПОЛНЫХ тёзок.
// UPD: вуз тоже)))
func (s *Storage) SaveAndGetTeacher(ctx context.Context, name string, cathedra models.Cathedra) models.Teacher {
	teacherQuery := "SELECT id FROM teachers WHERE name=$1;"
	var exists bool
	var teacher = models.Teacher{
		Name:      name,
		Cathedras: []models.Cathedra{cathedra},
	}
	rows, err := s.pool.QueryContext(ctx, teacherQuery, name)
	if err != nil {
		log.Printf("Ошибка при получения айди учителя: %s", err.Error())
		return teacher
	}
	defer rows.Close()

	// Если препод уже есть
	for rows.Next() {
		exists = true
		if err := rows.Scan(&teacher.Id); err != nil {
			log.Printf("Ошибка при получения айди учителя: %s", err.Error())
			return teacher
		}
		cathedraQuery := "INSERT INTO m2m_teachers_cathedras (teacher_id, cathedra_id) VALUES ($1, $2);"
		_, err = s.pool.ExecContext(ctx, cathedraQuery, teacher.Id, cathedra.Id)
		if err != nil {
			return teacher
		}
	}

	// Если такого преподавателя в принципе не существует
	if !exists {
		insQuery := "INSERT INTO teachers (name) VALUES ($1) RETURNING id;"
		insRows, err := s.pool.QueryContext(ctx, insQuery, name)
		pqErr, ok := err.(*pq.Error)
		if ok && string(pqErr.Code) == UniqueViolationCode {
			row := s.pool.QueryRowContext(ctx, teacherQuery, name)
			row.Scan(&teacher.Id)
		} else if err != nil {
			log.Printf("Ошибка при получении учителя: %s", err.Error())
			return teacher
		} else {
			defer insRows.Close()

			for insRows.Next() {
				if err := insRows.Scan(&teacher.Id); err != nil {
					log.Printf("Ошибка при получении учителя: %s", err.Error())
					return teacher
				}
			}
		}
		insQuery = "INSERT INTO m2m_teachers_cathedras (teacher_id, cathedra_id) VALUES ($1, $2);"
		_, err = s.pool.ExecContext(ctx, insQuery, teacher.Id, cathedra.Id)
		if err != nil {
			log.Printf("Ошибка при получении учителя: %s", err.Error())
			return teacher
		}
	}

	return teacher
}

func (s *Storage) SaveAndGetRoom(ctx context.Context, buildingNum int, room string) models.Room {
	var newRoom = models.Room{
		BuildingNum: buildingNum,
		Num:         room,
	}
	var exists bool
	const query = "SELECT id FROM rooms WHERE building_num=$1 AND num=$2;"
	rows, err := s.pool.QueryContext(ctx, query, buildingNum, room)
	if err != nil {
		log.Printf("Ошибка при получении  : %s", err.Error())
		return newRoom
	}

	for rows.Next() {
		exists = true
		if err := rows.Scan(&newRoom.Id); err != nil {
			// logrus
			return newRoom
		}
	}

	if !exists {
		const insQuery = "INSERT INTO rooms(building_num, num) VALUES ($1,$2) RETURNING id;"
		row := s.pool.QueryRowContext(ctx, insQuery, buildingNum, room)
		err := row.Scan(&newRoom.Id)
		pqErr, ok := err.(*pq.Error)
		if ok && string(pqErr.Code) == UniqueViolationCode {
			row := s.pool.QueryRowContext(ctx, query, buildingNum, room)
			row.Scan(&newRoom.Id)
		} else if err != nil {
			log.Printf("Ошибка при получении аудитории: %s", err.Error())
			return newRoom
		}
	}
	return newRoom
}

func (s *Storage) SaveLessons(ctx context.Context, lessons []models.Lesson) {
	const query = `INSERT INTO lessons(reaid, date, lesson_num, lesson_type_id, discipline_id, room_id, subgroup_num, cathedra_id, is_commission)
				   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				   ON CONFLICT (reaid, room_id) DO UPDATE
				   SET date=$2, lesson_num=$3, lesson_type_id=$4, discipline_id=$5, room_id=$6, subgroup_num=$7, cathedra_id=$8, is_commission=$9
				   RETURNING id;`
	stmt, err := s.pool.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Ошибка при сохранении пар: %s", err.Error())
		return
	}
	defer stmt.Close()

	for _, lesson := range lessons {
		row := stmt.QueryRow(
			lesson.ReaId,
			lesson.Date,
			lesson.
				LessonTime.Num,
			lesson.LessonType.Id,
			lesson.Discipline.Id,
			lesson.Room.Id,
			lesson.SubgroupNum,
			lesson.Cathedra.Id,
			lesson.IsCommission,
		)
		err := row.Scan(&lesson.Id)
		if err != nil {
			log.Printf("Ошибка при сохранении пар: %s", err.Error())
			return
		}
		const groupQuery = "INSERT INTO m2m_groups_lessons (group_id, lesson_id) VALUES ($1, $2);"
		const teacherQuery = "INSERT INTO m2m_teachers_lessons (teacher_id, lesson_id) VALUES ($1, $2);"

		groupStmt, err := s.pool.PrepareContext(ctx, groupQuery)
		if err != nil {
			log.Printf("Ошибка при сохранении пар: %s", err.Error())
			return
		}
		defer groupStmt.Close()

		for _, g := range lesson.Groups {
			groupStmt.Exec(g.Id, lesson.Id)
		}

		teacherStmt, err := s.pool.Prepare(teacherQuery)
		if err != nil {
			log.Printf("Ошибка при сохранении пар: %s", err.Error())
			return
		}
		defer teacherStmt.Close()
		for _, t := range lesson.Teachers {
			teacherStmt.Exec(t.Id, lesson.Id)
		}
	}
}

func (s *Storage) SaveOrUpdateGroups(ctx context.Context, groups []models.Group) {
	query := `INSERT INTO groups (name, faculty_id, course, education_type_id)
					  VALUES ($1,$2,$3,$4)
					  ON CONFLICT (name)
					  DO UPDATE SET name=$1, faculty_id=$2, course=$3, education_type_id=$4;`
	stmt, err := s.pool.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, group := range groups {
		_, err := stmt.ExecContext(ctx, group.Name, group.Faculty.Id, group.Course, group.EducationType.Id)
		if err != nil {
			log.Printf("Ошибка при сохранении групп: %s", err.Error())
			return
		}
	}
}

func (s *Storage) SaveAndGetFaculty(name string) models.Faculty {
	var faculty = models.Faculty{
		Name: name,
	}
	var exists bool
	const query = "SELECT id FROM faculties WHERE name = $1;"
	rows, err := s.pool.Query(query, name)
	if err != nil {
		log.Printf("Ошибка при сохранения факультета: %s", err.Error())
		return faculty
	}

	for rows.Next() {
		exists = true
		err := rows.Scan(&faculty.Id)
		if err != nil {
			return faculty
		}
	}
	if !exists {
		const insQuery = "INSERT INTO faculties(name) VALUES ($1) RETURNING id;"
		rows, err = s.pool.Query(insQuery, name)
		if err != nil {
			log.Printf("Ошибка при сохранения факультета: %s", err.Error())
			return faculty
		}
		for rows.Next() {
			err := rows.Scan(&faculty.Id)
			if err != nil {
				return faculty
			}
		}
	}
	return faculty
}

func (s *Storage) SaveAndGetEducationType(name string) models.EducationType {
	var educationType = models.EducationType{
		Name: name,
	}
	var exists bool
	const query = "SELECT id FROM education_types WHERE name = $1;"
	rows, err := s.pool.Query(query, name)
	if err != nil {
		log.Printf("Ошибка при сохранения формата обучения: %s", err.Error())
		return educationType
	}

	for rows.Next() {
		exists = true
		err := rows.Scan(&educationType.Id)
		if err != nil {
			return educationType
		}
	}
	if !exists {
		const insQuery = "INSERT INTO education_types(name) VALUES ($1) RETURNING id;"
		rows, err = s.pool.Query(insQuery, name)
		if err != nil {
			log.Printf("Ошибка при сохранения формата обучения: %s", err.Error())
			return educationType
		}
		for rows.Next() {
			err := rows.Scan(&educationType.Id)
			if err != nil {
				return educationType
			}
		}
	}
	return educationType
}
