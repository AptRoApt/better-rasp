package models

import "time"

// TODO: Нужно ли моделям копировать схему бд?

type Faculty struct {
	Id   int
	Name string
}

type EducationType struct {
	Id   int
	Name string
}

type Group struct {
	Id            int
	Name          string
	Faculty       Faculty
	Course        int
	EducationType EducationType
}

type Department struct {
	Id   int
	Name string
}

type Teacher struct {
	Id         int
	Name       string
	Department Department
}

type Building struct {
	Id      int
	Num     int
	Address string
}

type Room struct {
	Building Building
	Num      int
}

type LessonTime struct {
	Id        int
	StartTime time.Time
	EndTime   time.Time
}

type LessonType struct {
	Id   int
	Name string
}

type Discipline struct {
	Id   int
	Name string
}

type Lesson struct {
	Id             int
	Date           time.Time
	LessonTime     LessonTime
	LessonType     LessonType
	DisciplineName Discipline
	Teacher        Teacher
	Groups         []Group // Для лекций. В бд это отдельная таблица m2m.
	SubgroupNum    int     // Если занятие без подгрупп, то 0. Иначе 1 или 2.
}
