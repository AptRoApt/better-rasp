package storage

import (
	"better-rasp/internal/models"
	"context"
	"fmt"
	"log"
)

func (s *Storage) GetAllGroups(ctx context.Context) []models.Group {
	const query = `SELECT groups.id, groups.name, faculty_id, f.name, course, education_type_id, et.name FROM groups
				   JOIN faculties f ON faculty_id = f.id
				   JOIN education_types et ON education_type_id = et.id;`
	rows, err := s.pool.QueryContext(ctx, query)
	if err != nil {
		log.Println(fmt.Sprintf("Ошибка при получении всех групп из бд: %s", err.Error()))
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
			log.Println(fmt.Sprintf("Ошибка при получении всех групп из бд: %s", err.Error()))
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
		return lessonType
	}
	defer rows.Close()

	for rows.Next() {
		exists = true
		if err := rows.Scan(&lessonType.Id, &lessonType.Name); err != nil {
			return lessonType
		}
	}

	if !exists {
		insQuery := "INSERT INTO lesson_types (name) VALUES ($1) RETURNING id;"
		insRows, err := s.pool.QueryContext(ctx, insQuery, name)
		if err != nil {
			return lessonType
		}
		defer insRows.Close()

		for insRows.Next() {
			if err := insRows.Scan(&lessonType.Id); err != nil {
				return lessonType
			}
		}
		return lessonType
	}

	return lessonType

}

func (s *Storage) GetDisciplineByName(ctx context.Context, name string) models.Discipline {
	query := "SELECT id, name FROM disciplines WHERE name=$1;"
	var exists bool
	var discipline = models.Discipline{
		Name: name,
	}
	rows, err := s.pool.QueryContext(ctx, query, name)
	if err != nil {
		return discipline
	}
	defer rows.Close()

	for rows.Next() {
		exists = true
		if err := rows.Scan(&discipline.Id, &discipline.Name); err != nil {
			return discipline
		}
	}

	if !exists {
		insQuery := "INSERT INTO disciplines (name) VALUES ($1) RETURNING id;"
		insRows, err := s.pool.QueryContext(ctx, insQuery, name)
		if err != nil {
			return discipline
		}
		defer insRows.Close()

		for insRows.Next() {
			if err := insRows.Scan(&discipline.Id); err != nil {
				return discipline
			}
		}

		return discipline
	}

	return discipline
}

func (s *Storage) SaveAndGetCathedraByName(ctx context.Context, name string) models.Cathedra {
	query := "SELECT id, name FROM cathedras WHERE name=$1;"
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
		if err := rows.Scan(&cathedra.Id, &cathedra.Name); err != nil {
			return cathedra
		}
	}

	if !exists {
		insQuery := "INSERT INTO cathedras (name) VALUES ($1) RETURNING id;"
		insRows, err := s.pool.QueryContext(ctx, insQuery, name)
		if err != nil {
			log.Printf("Ошибка при получении кафедры: %s", err.Error())
			return cathedra
		}
		defer insRows.Close()

		for insRows.Next() {
			if err := insRows.Scan(&cathedra.Id); err != nil {
				log.Printf("Ошибка при получении кафедры: %s", err.Error())
				return cathedra
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
		return teacher
	}
	defer rows.Close()

	// Если препод уже есть
	for rows.Next() {
		exists = true
		if err := rows.Scan(&teacher.Id); err != nil {
			// logrus
			return teacher
		}
		cathedraQuery := "INSERT INTO m2m_teachers_cathedras (teacher_id, cathedra_id) VALUES ($1, $2);"
		_, err = s.pool.ExecContext(ctx, cathedraQuery, teacher.Id, cathedra.Id)
		if err != nil {
			// Ну, если он уже есть, то откатится, чё.
			return teacher
		}
	}

	// Если такого преподавателя в принципе не существует
	if !exists {
		insQuery := "INSERT INTO teachers (name) VALUES ($1) RETURNING id;"
		insRows, err := s.pool.QueryContext(ctx, insQuery, name)
		if err != nil {
			return teacher
		}
		defer insRows.Close()

		for insRows.Next() {
			if err := insRows.Scan(&teacher.Id); err != nil {
				//logrus
				return teacher
			}
		}

		insQuery = "INSERT INTO m2m_teachers_cathedras (teacher_id, cathedra_id) VALUES ($1, $2);"
		_, err = s.pool.ExecContext(ctx, insQuery, teacher.Id)
		if err != nil {
			//logrus
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
		if err != nil {
			//log
			return newRoom
		}
	}
	return newRoom
}

func (s *Storage) SaveLessons(ctx context.Context, lessons []models.Lesson) {
	const query = `INSERT INTO lessons(id, date, lesson_num, lesson_type_id, discipline_id, room_id, subgroup_num, cathedra_id, is_commission)
				   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				   ON CONFLICT (id) DO UPDATE
				   SET date=$2, lesson_num=$3, lesson_type_id=$4, discipline_id=$5, room_id=$6, subgroup_num=$7, cathedra_id=$8, is_commission=$9;`
	stmt, err := s.pool.PrepareContext(ctx, query)
	if err != nil {
		log.Println(fmt.Sprintf("Ошибка при сохранении пар: %s", err.Error()))
	}
	defer stmt.Close()

	for _, lesson := range lessons {
		_, err := stmt.ExecContext(ctx,
			lesson.Id,
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
		if err != nil {
			log.Println(fmt.Sprintf("Ошибка при сохранении пар: %s", err.Error()))
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
