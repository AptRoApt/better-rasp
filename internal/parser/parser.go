package parser

import (
	"better-rasp/internal/models"
	"better-rasp/internal/storage"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-co-op/gocron/v2"
)

type Parser struct {
	client    *http.Client
	storage   *storage.Storage
	scheduler gocron.Scheduler
}

type headerTransport struct {
}

func (ht headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	if req.Method == "POST" {
		req.Header.Set("User-Agent", "StudyProject (contact: ampore@mail.ru)")
	}
	return http.DefaultTransport.RoundTrip(req)
}

func New(s *storage.Storage) Parser {
	sr, err := gocron.NewScheduler()
	if err != nil {
		panic("Ошибка при создании планировщика.")
	}
	p := Parser{
		client:    &http.Client{Transport: headerTransport{}},
		storage:   s,
		scheduler: sr,
	}

	_, err = p.scheduler.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(2, 0, 0),
			),
		),
		gocron.NewTask(
			p.GetSchedule,
		),
	)
	if err != nil {
		panic("Ошибка при создании таски.")
	}
	return p
}

func (p *Parser) Start() {
	p.GetGroups()
	go p.GetSchedule()
	p.scheduler.Start()
}

func (p *Parser) Stop() {
	p.scheduler.StopJobs()
}

func (p *Parser) getElementsList(elementName string, data url.Values) ([]string, error) {
	resp, err := p.client.PostForm("https://rasp.rea.ru/Schedule/Navigator", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка при запросе к реа")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	options := doc.Find("select[name=\"" + elementName + "\"] option").FilterFunction(func(i int, s *goquery.Selection) bool {
		return i > 0
	})

	var result []string
	options.Each(func(i int, s *goquery.Selection) {
		result = append(result, s.Text())
	})

	return result, nil
}

func (p *Parser) getFacultiesGroups(faculty string) ([]models.Group, error) {
	data := url.Values{
		"Faculty":      {"na"},
		"Course":       {"na"},
		"Type":         {"na"},
		"ChangedNode":  {"na"},
		"ChangedValue": {"na"},
	}

	data.Set("Faculty", faculty)
	data.Set("ChangedNode", "Faculty")
	data.Set("ChangeValue", faculty)

	courses, err := p.getElementsList("Course", data)
	if err != nil {
		return nil, err
	}

	var facultyGroups []models.Group
	for _, course := range courses {
		data.Set("Course", course)
		data.Set("ChangedNode", "Course")
		data.Set("ChangeValue", course)

		types, err := p.getElementsList("Type", data)
		if err != nil {
			return nil, nil
		}
		for _, edType := range types {
			data.Set("Type", edType)
			data.Set("ChangedNode", "Type")
			data.Set("ChangeValue", edType)

			groups, err := p.getElementsList("Group", data)
			courseNum, _ := strconv.Atoi(string(course[0]))
			for _, group := range groups {
				facultyGroups = append(
					facultyGroups,
					models.Group{
						Name:          group,
						Faculty:       p.storage.SaveAndGetFaculty(faculty),
						Course:        courseNum,
						EducationType: p.storage.SaveAndGetEducationType(edType),
					},
				)
			}
			if err != nil {
				return nil, nil
			}
		}
	}
	return facultyGroups, nil
}

func (p *Parser) GetGroups() {
	// Получаем список факультетов
	resp, err := p.client.Get("https://rasp.rea.ru/Schedule/Navigator")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	options := doc.Find("select[name=\"Faculty\"] option").FilterFunction(func(i int, s *goquery.Selection) bool {
		return i > 0
	})

	var faculties []string
	options.Each(func(i int, s *goquery.Selection) {
		faculties = append(faculties, s.Text())
	})

	// Параллельно собираем список групп
	var groups []models.Group
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, faculty := range faculties {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			facultyGroups, err := p.getFacultiesGroups(f)
			if err != nil {
				log.Println(err)
				return
			}
			mutex.Lock()
			groups = append(groups, facultyGroups...)
			mutex.Unlock()
		}(faculty)
	}

	wg.Wait()

	for _, faculty := range faculties {
		facultyGroups, err := p.getFacultiesGroups(faculty)
		if err != nil {
			log.Println(err)
			return
		}
		groups = append(groups, facultyGroups...)
	}

	p.storage.SaveOrUpdateGroups(context.TODO(), groups)
}

func (p *Parser) getLesson(group models.Group, date string, timeslot int) []models.Lesson {
	var targetLink = fmt.Sprintf(
		"https://rasp.rea.ru/Schedule/GetDetails?selection=%s&date=%s&timeSlot=%v",
		strings.ToLower(group.Name),
		date,
		timeslot,
	)

	resp, err := p.client.Get(targetLink)
	if err != nil {
		//log somehow
		return nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		//log
		return nil
	}

	var lessons = make([]models.Lesson, 0, 2)

	elements := doc.Find(".element-info-body")
	withSubgroups := elements.Size() == 2

	elements.Each(
		func(i int, s *goquery.Selection) {
			//Получаем дату
			date, err := time.Parse("02.01.2006", date)
			if err != nil {
				//log
				return
			}

			var isCommission bool
			var lessonType string
			// Получаем тип занятия и есть ли комиссия
			s.Find("strong").Each(
				func(i int, s *goquery.Selection) {
					if i > 0 {
						isCommission = true
						return
					}
					lessonType = s.Text()
				})
			// Получаем номер подгруппы, если она есть
			var subgroupNum int
			if withSubgroups {
				subgroupNum = i + 1
			}
			// Парсим кафедру
			var cathedraName string
			cathedraStartIndex := strings.LastIndex(s.Text(), "(") + 1
			cathedraEndIndex := strings.LastIndex(s.Text(), ")")
			if cathedraEndIndex == -1 {
				// У преподавателей просто нет кафедры
				cathedraName = "нет"
			} else {
				cathedraName = s.Text()[cathedraStartIndex:cathedraEndIndex]
			}
			cathedra := p.storage.SaveAndGetCathedraByName(context.TODO(), cathedraName)
			// И корпус
			buildingIndex := strings.Index(s.Text(), "корпус") - 2
			buildingNum, err := strconv.Atoi(s.Text()[buildingIndex : buildingIndex+1])
			if err != nil {
				//log
				return
			}
			// И номер аудитории
			delta := strings.Index(s.Text()[buildingIndex+25:], "\n")
			room := s.Text()[buildingIndex+25 : buildingIndex+25+delta] // Там может быть Вебинар, так что это должна быть строка
			// И айди пары.
			id, err := strconv.Atoi(s.Find(".task-id-display").Text()[5:])
			if err != nil {
				//logrus
				return
			}

			// Получаем учителей(я)
			var teachers = make([]models.Teacher, 0, 1)
			s.Find("a").Each(
				func(i int, s *goquery.Selection) {
					teachers = append(teachers, p.storage.SaveAndGetTeacher(
						context.TODO(),
						s.Text()[7:],
						cathedra,
					))
				})

			var lesson = models.Lesson{
				ReaId:        id,
				Date:         date,
				LessonNum:    timeslot,
				LessonType:   p.storage.GetLessonTypeByName(context.TODO(), lessonType),
				Discipline:   p.storage.GetDisciplineByName(context.TODO(), s.Find("h5").Text()),
				Room:         p.storage.SaveAndGetRoom(context.TODO(), buildingNum, room),
				Teachers:     teachers,
				Groups:       []models.Group{group},
				SubgroupNum:  subgroupNum,
				Cathedra:     cathedra,
				IsCommission: isCommission,
			}
			lessons = append(lessons, lesson)
		})

	return lessons
}

func (p *Parser) getScheduleForGroup(group models.Group, weekNum int) {
	var targetLink = fmt.Sprintf(
		"https://rasp.rea.ru/Schedule/ScheduleCard?selection=%s&weekNum=%v&catfilter=0",
		strings.ToLower(group.Name),
		weekNum,
	)
	resp, err := p.client.Get(targetLink)
	if err != nil {
		log.Printf("Ошибка при получении основной страницы расписания: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Ошибка при парсинге основной страницы расписания: %s", err.Error())
		return
	}

	var lessons []models.Lesson

	doc.Find("[class*=slot]:not([class*=empty])").Each(
		func(i int, s *goquery.Selection) {
			v, e := s.Find("[onclick]").Attr("onclick")
			if e {
				date := v[21:31]
				timeslot, err := strconv.Atoi(v[35:36])
				if err != nil {
					//logrus
					return
				}
				lessons = append(lessons, p.getLesson(group, date, timeslot)...)
			}
		})

	p.storage.SaveLessons(context.TODO(), lessons)
}

func (p *Parser) GetSchedule() {
	_, weekNum := time.Now().ISOWeek()
	groups := p.storage.GetAllGroups(context.TODO())
	weekNum += 18
	var wg sync.WaitGroup
	for _, group := range groups {
		wg.Add(1)
		go func(gr models.Group, wN int) {
			defer wg.Done()
			p.getScheduleForGroup(gr, wN)
		}(group, weekNum)
		wg.Add(1)
		go func(gr models.Group, wN int) {
			defer wg.Done()
			p.getScheduleForGroup(gr, wN)
		}(group, (weekNum)%53+1)
		wg.Add(1)
		go func(gr models.Group, wN int) {
			defer wg.Done()
			p.getScheduleForGroup(gr, wN)
		}(group, (weekNum)%53+2)
		wg.Wait()
	}
	log.Println("Получено всё расписание!")
}
