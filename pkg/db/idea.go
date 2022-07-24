package db

import (
	"gorm.io/gorm"
)

type Idea struct {
	gorm.Model
	User        string
	Description string
}
