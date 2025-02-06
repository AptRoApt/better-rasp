package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func EnvConfig() Config {
	host := os.Getenv("db_host")

	portString := os.Getenv("db_port")

	port, _ := strconv.Atoi(portString)

	user := os.Getenv("db_user")

	password := os.Getenv("db_password")

	database := os.Getenv("db_database")

	return Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
}

func (c Config) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Database)
}

type Storage struct {
	config Config
	pool   *sql.DB
	log    *logrus.Logger
}

func New(c Config, l *logrus.Logger) Storage {
	if c.Host == "" {
		l.Info("переменная Host не установлена. установлено дефолтное значение")
		c.Host = "localhost"
	}
	if c.Port == 0 {
		l.Info("переменная Port не установлена. установлено дефолтное значение")
		c.Port = 5432
	}
	if c.User == "" {
		panic("Не задан пользователь СУБД!")
	}
	if c.Password == "" {
		panic("Не задан пароль пользователя СУБД!")
	}
	if c.Database == "" {
		panic("Не задано имя БД!")
	}

	if c.Host == "" || c.Port == 0 || c.User == "" || c.Password == "" || c.Database == "" {
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
	tx, err := s.pool.Begin()
	if err != nil {
		return fmt.Errorf("ошибка при создании транзакции: %w", err)
	}

	query, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка при чтении файла с миграцией: %w", err)
	}
	_, err = tx.Exec(string(query))
	if err != nil {
		return fmt.Errorf("ошибка при проведении транзакции: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("ошибка при коммите транзакции: %w", err)
	}
	return nil
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
