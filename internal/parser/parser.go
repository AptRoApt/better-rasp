package parser

import (
	"better-rasp/internal/models"
	"better-rasp/internal/storage"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

func New(s *storage.Storage) Parser {
	sr, err := gocron.NewScheduler()
	if err != nil {
		panic("Ошибка при создании планировщика.")
	}
	p := Parser{
		client:    http.DefaultClient,
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
		return nil, fmt.Errorf("Ошибка при запросе к реа")
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

func (p *Parser) getFacultiesGroups(faculty string, data url.Values) ([]models.Group, error) {
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
			for _, group := range groups {
				facultyGroups = append(
					facultyGroups,
					models.Group{
						Name:          group,
						Faculty:       models.Faculty{Name: faculty},
						Course:        int(course[0]),
						EducationType: models.EducationType{Name: edType},
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
	data := url.Values{
		"Faculty":      {"na"},
		"Course":       {"na"},
		"Type":         {"na"},
		"ChangedNode":  {"na"},
		"ChangedValue": {"na"},
	}
	faculties, err := p.getElementsList("Faculty", data)
	if err != nil {
		log.Println(err)
		return
	}
	// Параллельно собираем список групп
	var groups []models.Group
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, faculty := range faculties {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			facultyGroups, err := p.getFacultiesGroups(f, data)
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

	err = p.storage.AddGroups(context.TODO(), groups)
	if err != nil {
		log.Println(err)
		return
	}
}

func (p *Parser) getScheduleForGroup(group string, weekNum int) {

}

func (p *Parser) GetSchedule() {
	_, weekNum := time.Now().ISOWeek()
	groups := p.storage.GetAllGroups()
	for _, group := range groups {
		p.getScheduleForGroup(group.Name, weekNum)
		p.getScheduleForGroup(group.Name, weekNum%53+1)
		p.getScheduleForGroup(group.Name, weekNum%53+2)
	}
}
