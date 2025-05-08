// TODO: Вынести запросы в отдельный модуль internal/storage/queries
package storage

import (
	"better-rasp/internal/models"
	"context"
	"log"
)

// Получить список всех групп
func (s *Storage) GetAllGroups(ctx context.Context) []models.Group {
	const query = `SELECT groups.id, groups.name, faculty_id, f.name, course, education_type_id, et.name FROM groups
				   JOIN faculties f ON faculty_id = f.id
				   JOIN education_types et ON education_type_id = et.id;`
	rows, err := s.pool.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Ошибка при получении всех групп из бд: %s", err.Error())
		return nil
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
			continue
		}
		groups = append(groups, group)
	}
	return groups
}

func (s *Storage) UpsertLessonType(ctx context.Context, name string) models.LessonType {
	query := `INSERT INTO lesson_types (name) VALUES ($1)
			  ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			  RETURNING id, name;`

	lessonType := models.LessonType{Name: name}
	err := s.pool.QueryRowContext(ctx, query, name).Scan(&lessonType.Id, &lessonType.Name)
	if err != nil {
		log.Printf("Ошибка при получении/создании типа занятия: %s", err.Error())
	}
	return lessonType
}

func (s *Storage) UpsertDiscipline(ctx context.Context, name string) models.Discipline {
	query := `INSERT INTO disciplines (name) VALUES ($1)
			  ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			  RETURNING id;`

	discipline := models.Discipline{Name: name}
	err := s.pool.QueryRowContext(ctx, query, name).Scan(&discipline.Id)
	if err != nil {
		log.Printf("Ошибка при получении/создании дисциплины: %s", err.Error())
	}
	return discipline
}

func (s *Storage) SaveAndGetCathedraByName(ctx context.Context, name string) models.Cathedra {
	query := `INSERT INTO cathedras (name) VALUES ($1)
			  ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			  RETURNING id;`

	cathedra := models.Cathedra{Name: name}
	err := s.pool.QueryRowContext(ctx, query, name).Scan(&cathedra.Id)
	if err != nil {
		log.Printf("Ошибка при получении/создании кафедры: %s", err.Error())
	}
	return cathedra
}

// я ОЧЕНЬ надеюсь, что у нас нет ПОЛНЫХ тёзок.
// UPD: вуз тоже)))
func (s *Storage) UpsertTeacher(ctx context.Context, name string, cathedra models.Cathedra) models.Teacher {
	// Транзакция для атомарного создания учителя и связи с кафедрой
	tx, err := s.pool.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Ошибка при создании транзакции: %s", err.Error())
		return models.Teacher{Name: name, Cathedras: []models.Cathedra{cathedra}}
	}
	defer tx.Rollback()

	// Вставляем или получаем ID преподавателя
	teacherQuery := `INSERT INTO teachers (name) VALUES ($1)
					 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
					 RETURNING id;`

	var teacherId int
	err = tx.QueryRowContext(ctx, teacherQuery, name).Scan(&teacherId)
	if err != nil {
		log.Printf("Ошибка при получении/создании учителя: %s", err.Error())
		return models.Teacher{Name: name, Cathedras: []models.Cathedra{cathedra}}
	}

	// Добавляем связь преподавателя с кафедрой, если её нет
	cathedraQuery := `INSERT INTO m2m_teachers_cathedras (teacher_id, cathedra_id) 
					  VALUES ($1, $2)
					  ON CONFLICT (teacher_id, cathedra_id) DO NOTHING;`
	_, err = tx.ExecContext(ctx, cathedraQuery, teacherId, cathedra.Id)
	if err != nil {
		log.Printf("Ошибка при связывании учителя с кафедрой: %s", err.Error())
		return models.Teacher{Name: name, Cathedras: []models.Cathedra{cathedra}}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Ошибка при коммите транзакции: %s", err.Error())
		return models.Teacher{Name: name, Cathedras: []models.Cathedra{cathedra}}
	}

	return models.Teacher{
		Id:        teacherId,
		Name:      name,
		Cathedras: []models.Cathedra{cathedra},
	}
}

func (s *Storage) UpsertRoom(ctx context.Context, buildingNum int, room string) models.Room {
	query := `INSERT INTO rooms (building_num, num) VALUES ($1, $2)
			  ON CONFLICT (building_num, num) DO UPDATE 
			  SET building_num = EXCLUDED.building_num, num = EXCLUDED.num
			  RETURNING id;`

	newRoom := models.Room{
		BuildingNum: buildingNum,
		Num:         room,
	}

	err := s.pool.QueryRowContext(ctx, query, buildingNum, room).Scan(&newRoom.Id)
	if err != nil {
		log.Printf("Ошибка при получении/создании аудитории: %s", err.Error())
	}
	return newRoom
}

func (s *Storage) SaveLessons(ctx context.Context, lessons []models.Lesson) {
	tx, err := s.pool.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Ошибка при создании транзакции: %s", err.Error())
		return
	}
	defer tx.Rollback()

	// Подготовка основного запроса для уроков
	const query = `INSERT INTO lessons(reaid, date, lesson_num, lesson_type_id, discipline_id, room_id, subgroup_num, cathedra_id, is_commission)
				   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				   ON CONFLICT (reaid, room_id) DO UPDATE
				   SET date=$2, lesson_num=$3, lesson_type_id=$4, discipline_id=$5, room_id=$6, subgroup_num=$7, cathedra_id=$8, is_commission=$9
				   RETURNING id;`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Ошибка при подготовке запроса для сохранения пар: %s", err.Error())
		return
	}
	defer stmt.Close()

	// Подготовка запросов для связей с группами и преподавателями
	const groupQuery = `INSERT INTO m2m_groups_lessons (group_id, lesson_id) 
					   VALUES ($1, $2)
					   ON CONFLICT (group_id, lesson_id) DO NOTHING;`

	const teacherQuery = `INSERT INTO m2m_teachers_lessons (teacher_id, lesson_id) 
						  VALUES ($1, $2)
						  ON CONFLICT (teacher_id, lesson_id) DO NOTHING;`

	groupStmt, err := tx.PrepareContext(ctx, groupQuery)
	if err != nil {
		log.Printf("Ошибка при подготовке запроса для связи с группами: %s", err.Error())
		return
	}
	defer groupStmt.Close()

	teacherStmt, err := tx.PrepareContext(ctx, teacherQuery)
	if err != nil {
		log.Printf("Ошибка при подготовке запроса для связи с преподавателями: %s", err.Error())
		return
	}
	defer teacherStmt.Close()

	for _, lesson := range lessons {
		var lessonId int
		err := stmt.QueryRowContext(ctx,
			lesson.ReaId,
			lesson.Date,
			lesson.LessonNum,
			lesson.LessonType.Id,
			lesson.Discipline.Id,
			lesson.Room.Id,
			lesson.SubgroupNum,
			lesson.Cathedra.Id,
			lesson.IsCommission,
		).Scan(&lessonId)

		if err != nil {
			log.Printf("Ошибка при сохранении пары (ID: %d): %s", lesson.ReaId, err.Error())
			continue
		}

		// Связь с группами
		for _, g := range lesson.Groups {
			_, err := groupStmt.ExecContext(ctx, g.Id, lessonId)
			if err != nil {
				log.Printf("Ошибка при связывании пары с группой: %s", err.Error())
			}
		}

		// Связь с преподавателями
		for _, t := range lesson.Teachers {
			_, err := teacherStmt.ExecContext(ctx, t.Id, lessonId)
			if err != nil {
				log.Printf("Ошибка при связывании пары с преподавателем: %s", err.Error())
			}
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Ошибка при коммите транзакции: %s", err.Error())
	}
}

func (s *Storage) SaveOrUpdateGroups(ctx context.Context, groups []models.Group) {
	query := `INSERT INTO groups (name, faculty_id, course, education_type_id)
			  VALUES ($1, $2, $3, $4)
			  ON CONFLICT (name)
			  DO UPDATE SET faculty_id=$2, course=$3, education_type_id=$4;`

	stmt, err := s.pool.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Ошибка при подготовке запроса для сохранения групп: %s", err.Error())
		return
	}
	defer stmt.Close()

	for _, group := range groups {
		_, err := stmt.ExecContext(ctx, group.Name, group.Faculty.Id, group.Course, group.EducationType.Id)
		if err != nil {
			log.Printf("Ошибка при сохранении группы %s: %s", group.Name, err.Error())
		}
	}
}

func (s *Storage) SaveAndGetFaculty(ctx context.Context, name string) models.Faculty {
	query := `INSERT INTO faculties (name) VALUES ($1)
			  ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			  RETURNING id;`

	faculty := models.Faculty{Name: name}
	err := s.pool.QueryRowContext(ctx, query, name).Scan(&faculty.Id)
	if err != nil {
		log.Printf("Ошибка при получении/создании факультета: %s", err.Error())
	}
	return faculty
}

func (s *Storage) SaveAndGetEducationType(ctx context.Context, name string) models.EducationType {
	query := `INSERT INTO education_types (name) VALUES ($1)
			  ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			  RETURNING id;`

	educationType := models.EducationType{Name: name}
	err := s.pool.QueryRowContext(ctx, query, name).Scan(&educationType.Id)
	if err != nil {
		log.Printf("Ошибка при получении/создании формата обучения: %s", err.Error())
	}
	return educationType
}
