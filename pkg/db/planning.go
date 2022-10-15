package db

import (
	"time"

	"gorm.io/gorm"
)

type Plan struct {
	gorm.Model
	User        string
	Description string
	ChannelID   string
	Annoying    bool // should the bot bother the user a lot...
	Print       bool // should the alert be sent to the printer

	Start       time.Time
	SnoozedTill time.Time
}
