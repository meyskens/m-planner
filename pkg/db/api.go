package db

import (
	"gorm.io/gorm"
)

type ApiToken struct {
	gorm.Model
	User  string
	Token string
}
