package schedule

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
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
}

func New(c Config) Storage {
	if c.Host == "" || c.Port == 0 || c.User == "" || c.Password == "" || c.DBName == "" {
		panic("Empty configuration fields")
	}

	s := Storage{
		config: c,
	}

	var err error
	s.pool, err = sql.Open("postgres", c.String())
	if err != nil {
		panic(err.Error())
	}

	// Проверка соединения
	err = s.pool.Ping()
	if err != nil {
		panic(err.Error())
	}

	return s
}

func (s *Storage) Close() {
	// тут не должно быть что-то умнее?
	s.pool.Close()
}
