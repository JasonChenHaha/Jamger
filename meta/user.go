package jmeta

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string `gorm:"type:varchar(20);not null"`
}
