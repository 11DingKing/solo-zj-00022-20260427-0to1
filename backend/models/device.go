package models

import (
	"time"

	"gorm.io/gorm"
)

type DeviceStatus string

const (
	DeviceStatusActive      DeviceStatus = "active"
	DeviceStatusMaintenance DeviceStatus = "maintenance"
	DeviceStatusScrapped    DeviceStatus = "scrapped"
)

type Device struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	DeviceCode        string         `json:"device_code" gorm:"uniqueIndex;size:50;not null"`
	Name              string         `json:"name" gorm:"size:100;not null"`
	Model             string         `json:"model" gorm:"size:100"`
	Location          string         `json:"location" gorm:"size:200"`
	PurchaseDate      *time.Time     `json:"purchase_date" gorm:"type:date"`
	WarrantyExpireDate *time.Time   `json:"warranty_expire_date" gorm:"type:date"`
	Status            DeviceStatus   `json:"status" gorm:"type:enum('active','maintenance','scrapped');default:'active';not null"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}
