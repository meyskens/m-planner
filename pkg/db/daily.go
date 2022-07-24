package db

import (
	"gorm.io/gorm"
)

type Daily struct {
	gorm.Model
	User        string
	Description string
	Hour        int
	Minute      int
	ChannelID   string
}
