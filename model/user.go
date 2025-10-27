package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Nickname string `gorm:"type:varchar(16)"`
	Password string `gorm:"type:varchar(256)"`
	Email    string `gorm:"unique"` 
	Volume   uint64 `gorm:"default:5"`//5GB
	Files    []File `gorm:"many2many:user_file"`
}
