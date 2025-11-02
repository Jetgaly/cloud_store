package model

import "time"

/*
status:
0:pending
1:over
2:cancel
*/
type VolumeOpLog struct {
	ID        int64  `gorm:"primaryKey"`
	UploadId  int64  `gorm:"unique,index:idx_uploadid"`
	UserID    int64  `gorm:"not null;index:idx_user_status"`              // 用户ID
	FileHash  string `gorm:"not null;size:64;index:idx_hash"`             // 文件hash
	FileSize  uint64 `gorm:"not null"`                                    // 文件大小
	Status    int8   `gorm:"type:TINYINT;not null;index:idx_user_status"` // 状态
	FileName  string `gorm:"type:varchar(128);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
