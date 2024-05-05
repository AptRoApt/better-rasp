package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func (c Config) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}

type Storage struct {
	config Config
	pool   *sql.DB
	log    *logrus.Logger
}

func New(c Config, l *logrus.Logger) Storage {
	if c.Host == "" || c.Port == 0 || c.User == "" || c.Password == "" || c.DBName == "" {
		panic("Пустые поля в конфиге для бд недопустимы.")
	}

	s := Storage{
		config: c,
		log:    l,
	}

	var err error
	s.pool, err = sql.Open("postgres", c.String())
	if err != nil {
		panic(fmt.Sprintf("ошибка при инициализации соединения с бд: %s", err.Error()))
	}

	// Проверка соединения
	err = s.pool.Ping()
	if err != nil {
		panic(fmt.Sprintf("ошибка при проверке соединения с бд: %s", err.Error()))
	}

	err = s.InitDatabase()
	if err != nil {
		panic(fmt.Sprintf("ошибка при инициализации бд: %s", err.Error()))
	}

	return s
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) executeMigrationScript(filename string) error {
	filePath := filepath.Join("migrations", filename)
	query, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка при чтении файла с миграцией: %w", err)
	}

	_, err = s.pool.Exec(string(query))
	return errors.Join(errors.New("ошибка при выполнении миграции"), err)
}

func (s *Storage) InitDatabase() error {
	return s.executeMigrationScript("initdb.sql")
}

func (s *Storage) TruncateDatabase() error {
	return s.executeMigrationScript("truncatedb.sql")
}

func (s *Storage) DropDatabase() error {
	return s.executeMigrationScript("droptables.sql")
}
