package db

import (
	"time"

	"gorm.io/gorm"
)

type RecycleReminder struct {
	gorm.Model
	User      string
	ChannelID string

	StreetID     string
	PostalCodeID string
	HouseNumber  string
	Annoying     bool // should the bot bother the user a lot...

	LastRun time.Time
}
