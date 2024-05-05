package storage

import (
	"better-rasp/internal/models"
	"context"
)

func (s *Storage) GetDepartments(ctx context.Context) ([]models.Department, error) {
	s.log.Debug("Getting departments...")

	query := "SELECT id, name FROM departments;"
	rows, err := s.pool.QueryContext(ctx, query)
	if err != nil {
		s.log.WithError(err).Error("Error querying database")
		return nil, err
	}
	defer rows.Close()

	var departments []models.Department

	for rows.Next() {
		var department models.Department
		err := rows.Scan(&department.Id, &department.Name)
		if err != nil {
			s.log.WithError(err).Error("Error scanning row")
			return nil, err
		}
		departments = append(departments, department)
	}

	s.log.WithField("count", len(departments)).Debug("Retrieved departments")

	return departments, nil
}

func (s *Storage) GetTeachers(ctx context.Context, department models.Department) ([]models.Teacher, error) {
	s.log.Debug("Getting teachers...")

	query := "SELECT t_id, t_name FROM teachers_view WHERE dp_id = $1;"
	rows, err := s.pool.QueryContext(ctx, query, department.Id)
	if err != nil {
		s.log.WithError(err).Error("Error querying database")
		return nil, err
	}
	defer rows.Close()

	var teachers []models.Teacher

	for rows.Next() {
		var teacher = models.Teacher{
			Department: department,
		}
		err := rows.Scan(&teacher.Id, &teacher.Name)
		if err != nil {
			s.log.WithError(err).Error("Error scanning row")
			return nil, err
		}
		teachers = append(teachers, teacher)
	}

	s.log.WithField("count", len(teachers)).Debug("Retrieved teachers")

	return teachers, nil
}
