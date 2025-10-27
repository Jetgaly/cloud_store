package model

import "gorm.io/gorm"

type File struct{
	gorm.Model
	Name string `gorm:"type:varchar(128);not null"`
	Hash string `gorm:"type:char(64);uniqueIndex;not null"`
	Path string `gorm:"type:varchar(1024);not null"`
	Size uint64 `gorm:"not null"`
	Status int8 `gorm:"type:TINYINT;default:0;not null"`//0:可用
}