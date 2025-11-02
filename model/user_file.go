package model

import "time"

type UserFile struct {
	UserId    int64  `gorm:"primaryKey"`
	FileId    int64  `gorm:"primaryKey"`
	FileName  string `gorm:"type:varchar(512)"`
	User      User   `gorm:"foreignKey:UserId;references:ID"`
	File      File   `gorm:"foreignKey:FileId;references:ID"`
	CreatedAt time.Time
}
