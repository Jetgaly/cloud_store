package model

import "time"

type UserFile struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`     // 单独主键
	UserId    int64  `gorm:"not null;index:idx_user_file"` 
	FileId    int64  `gorm:"not null;index:idx_user_file"` 
	FileName  string `gorm:"type:varchar(512)"`
	User      User   `gorm:"foreignKey:UserId;references:ID"`
	File      File   `gorm:"foreignKey:FileId;references:ID"`
	CreatedAt time.Time
}
