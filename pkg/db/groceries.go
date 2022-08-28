package db

import (
	"gorm.io/gorm"
)

type Grocery struct {
	gorm.Model
	Item      string
	ChannelID string
}
