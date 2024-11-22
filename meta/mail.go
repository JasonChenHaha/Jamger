package jmeta

import (
	"gorm.io/gorm"
)

type Mail struct {
	gorm.Model
	Data string `gorm:"type:text;"`
}
