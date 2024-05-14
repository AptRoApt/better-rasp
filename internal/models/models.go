package models

import "time"

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

type Cathedra struct {
	Id   int
	Name string
}

type Teacher struct {
	Id        int
	Name      string
	Cathedras []Cathedra
}

type Room struct {
	Id          int
	BuildingNum int
	Num         string
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
	ReaId        int
	Id           int
	Date         time.Time
	LessonNum    int
	LessonType   LessonType
	Discipline   Discipline
	Room         Room
	Teachers     []Teacher // Может быть комиссия.
	Groups       []Group   // Для лекций. В бд это отдельная таблица m2m.
	SubgroupNum  int       // Если занятие без подгрупп, то 0. Иначе 1 или 2.
	Cathedra     Cathedra
	IsCommission bool
}
