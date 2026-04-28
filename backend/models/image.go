package models

import (
	"time"

	"gorm.io/gorm"
)

type ImageType string

const (
	ImageTypeBefore ImageType = "before"
	ImageTypeAfter  ImageType = "after"
)

type Image struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	WorkOrderID *uint          `json:"work_order_id" gorm:"index"`
	ImageType   ImageType      `json:"image_type" gorm:"type:enum('before','after');not null;default:'before'"`
	FilePath    string         `json:"file_path" gorm:"size:500;not null"`
	FileName    string         `json:"file_name" gorm:"size:255;not null"`
	FileSize    int            `json:"file_size"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
