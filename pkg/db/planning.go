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

	Start       time.Time
	SnoozedTill time.Time
}
