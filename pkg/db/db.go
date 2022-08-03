package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var tables = []interface{}{
	Idea{},
	Daily{},
	DailyReminder{},
	DailyReminderEvent{},
	Plan{},
	SentMessage{},
}

type Validatable interface {
	Validate() error
}

var ErrorNotFound = errors.New("not found")

type Connection struct {
	*gorm.DB
	sqlitePath string
	delete     bool
}

type ConnectionDetails struct {
	Host      string
	Port      int
	User      string
	Database  string
	Password  string
	EnableSSL bool
}

func NewConnection(details ConnectionDetails) (*Connection, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      false,       // Disable color
		},
	)

	var err error
	sslMode := "sslmode=disable"
	if details.EnableSSL {
		sslMode = ""
	}
	c, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s %s",
		details.Host, details.Port, details.User, details.Database, details.Password, sslMode)), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	return &Connection{c, "", false}, err
}

func (c *Connection) DoMigrate() error {
	for _, t := range tables {
		err := c.AutoMigrate(&t)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Connection) Close() error {
	if c.delete {
		return os.Remove(c.sqlitePath)
	}
	return nil
}
