package db

import (
	"gorm.io/gorm"
)

type PrintJob struct {
	gorm.Model
	User       string `json:"user"`
	EscposData []byte `json:"escposData"`
}
