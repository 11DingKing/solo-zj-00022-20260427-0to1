package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleEmployee   UserRole = "employee"
	RoleTechnician UserRole = "technician"
	RoleAdmin      UserRole = "admin"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string         `json:"-" gorm:"size:255;not null"`
	RealName  string         `json:"real_name" gorm:"size:50;not null"`
	Role      UserRole       `json:"role" gorm:"type:enum('employee','technician','admin');default:'employee';not null"`
	Status    UserStatus     `json:"status" gorm:"type:enum('active','inactive');default:'active';not null"`
	Phone     string         `json:"phone" gorm:"size:20"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" {
		return u.HashPassword()
	}
	return nil
}
