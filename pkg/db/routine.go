package db

import (
	"time"

	"gorm.io/gorm"
)

type Routine struct {
	gorm.Model
	User        string
	Description string
	ChannelID   string
	Print       bool // should the alert be sent to the printer
	Message     bool // should send discord message

	FunFact bool // should send a fun fact

	Reminders []RoutineReminder `gorm:"constraint:OnDelete:CASCADE;"`
}

type RoutineReminder struct {
	gorm.Model

	RoutineID uint
	Weekday   time.Weekday
	Hour      int
	Minute    int
}
