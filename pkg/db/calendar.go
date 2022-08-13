package db

import (
	"gorm.io/gorm"
)

type Calendar struct {
	gorm.Model
	User string
	Name string
	Link string
}
