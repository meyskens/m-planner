package db

import (
	"time"

	"gorm.io/gorm"
)

type Daily struct {
	gorm.Model
	User        string
	Description string
	ChannelID   string
	Annoying    bool // should the bot bother the user a lot...
	Print       bool // should the alert be sent to the printer

	Reminders []DailyReminder `gorm:"constraint:OnDelete:CASCADE;"`
}

type DailyReminder struct {
	gorm.Model

	DailyID uint
	Weekday time.Weekday
	Hour    int
	Minute  int
}

type DailyReminderEvent struct {
	gorm.Model

	DailyID     uint
	Daily       Daily
	Start       time.Time
	SnoozedTill time.Time

	SentMessages []SentMessage
}

type SentMessage struct {
	gorm.Model
	DailyReminderEventID uint
	MessageID            string
}
