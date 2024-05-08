package storage

import (
	"better-rasp/internal/models"
	"context"

	"github.com/sirupsen/logrus"
)

// TODO: Оптимизировать поиск группы с помощью индекса.

// TODO: Переписать всё к чертям. Зачем мне передавать всю структуру, если мне нужен лишь id?
func (s *Storage) GetGroupsByName(ctx context.Context, givenGroup string) ([]models.Group, error) {
	s.log.WithFields(logrus.Fields{"group": givenGroup}).Debug("Getting groups by name")

	query := "SELECT g_id, g_name, f_id, f_name, course, et_id, et_name FROM groups_view WHERE name LIKE '%$1';"
	rows, err := s.pool.QueryContext(ctx, query, givenGroup)
	if err != nil {
		s.log.WithError(err).Error("Error querying database")
		return nil, err
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
			s.log.WithError(err).Error("Error scanning row")
			return nil, err
		}
		groups = append(groups, group)
	}

	s.log.WithField("count", len(groups)).Debug("Retrieved groups")

	return groups, nil
}

func (s *Storage) GetFaculties(ctx context.Context) ([]models.Faculty, error) {
	s.log.Debug("Getting faculties")

	query := "SELECT id, name FROM faculties;"
	rows, err := s.pool.QueryContext(ctx, query)
	if err != nil {
		s.log.WithError(err).Error("Error querying database")
		return nil, err
	}
	defer rows.Close()

	var faculties []models.Faculty

	for rows.Next() {
		var faculty models.Faculty
		err := rows.Scan(&faculty.Id, &faculty.Name)
		if err != nil {
			s.log.WithError(err).Error("Error scanning row")
			return nil, err
		}
		faculties = append(faculties, faculty)
	}

	s.log.WithField("count", len(faculties)).Debug("Retrieved faculties")

	return faculties, nil
}

func (s *Storage) GetCourses(ctx context.Context, faculty models.Faculty) ([]int, error) {
	s.log.WithField("faculty", faculty.Name).Debug("Getting courses")

	query := "SELECT DISTINCT course FROM groups WHERE f_id=$1;"
	rows, err := s.pool.QueryContext(ctx, query, faculty.Id)
	if err != nil {
		s.log.WithError(err).Error("Error querying database")
		return nil, err
	}
	defer rows.Close()

	var courses []int

	for rows.Next() {
		var course int
		err := rows.Scan(&course)
		if err != nil {
			s.log.WithError(err).Error("Error scanning row")
			return nil, err
		}
		courses = append(courses, course)
	}
	s.log.WithField("count", len(courses)).Debug("Retrieved courses")
	return courses, nil
}

func (s *Storage) GetTypes(ctx context.Context, faculty models.Faculty, course int) ([]models.EducationType, error) {
	s.log.WithFields(logrus.Fields{"faculty": faculty.Name, "course": course}).Debug("Getting education types")

	query := "SELECT DISTINCT et_id, et_name FROM groups_view WHERE f_id=$1 AND course=$2;"
	rows, err := s.pool.QueryContext(ctx, query, faculty.Id, course)
	if err != nil {
		s.log.WithError(err).Error("Error querying database")
		return nil, err
	}
	defer rows.Close()

	var educationTypes []models.EducationType

	for rows.Next() {
		var educationType models.EducationType
		err := rows.Scan(&educationType.Id, &educationType.Name)
		if err != nil {
			s.log.WithError(err).Error("Error scanning row")
			return nil, err
		}
		educationTypes = append(educationTypes, educationType)
	}
	s.log.WithField("count", len(educationTypes)).Debug("Retrieved education types")
	return educationTypes, nil
}

func (s *Storage) GetGroups(ctx context.Context, faculty models.Faculty, course int, educationType models.EducationType) ([]models.Group, error) {
	s.log.WithFields(logrus.Fields{"faculty": faculty.Name, "course": course, "educationType": educationType.Id}).Debug("Getting groups")

	query := "SELECT g_id, g_name FROM groups WHERE f_id=$1 AND course=$2 AND et_id=$3;"
	rows, err := s.pool.QueryContext(ctx, query, faculty.Id, course, educationType.Id)
	if err != nil {
		s.log.WithError(err).Error("Error querying database")
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group

	for rows.Next() {
		var group = models.Group{
			Faculty:       faculty,
			Course:        course,
			EducationType: educationType,
		}
		err := rows.Scan(&group.Id, &group.Name)
		if err != nil {
			s.log.WithError(err).Error("Error scanning row")
			return nil, err
		}
		groups = append(groups, group)
	}
	s.log.WithField("count", len(groups)).Debug("Retrieved groups")
	return groups, nil
}

func (s *Storage) AddGroups(ctx context.Context, groups []models.Group) error {

}

func (s *Storage) GetAllGroups() []models.Group {

}
