package models

import (
	"time"

	"gorm.io/gorm"
)

type OperationType string

const (
	OperationCreate  OperationType = "create"
	OperationAssign  OperationType = "assign"
	OperationAccept  OperationType = "accept"
	OperationProcess OperationType = "process"
	OperationSubmit  OperationType = "submit"
	OperationConfirm OperationType = "confirm"
	OperationReject  OperationType = "reject"
)

type OperationLog struct {
	ID          uint             `json:"id" gorm:"primaryKey"`
	WorkOrderID uint             `json:"work_order_id" gorm:"not null;index"`
	UserID      uint             `json:"user_id" gorm:"not null;index"`
	Operation   OperationType    `json:"operation" gorm:"type:enum('create','assign','accept','process','submit','confirm','reject');not null"`
	OldStatus   *WorkOrderStatus `json:"old_status" gorm:"type:enum('pending_assign','assigned','processing','pending_confirm','closed')"`
	NewStatus   *WorkOrderStatus `json:"new_status" gorm:"type:enum('pending_assign','assigned','processing','pending_confirm','closed')"`
	Remark      string           `json:"remark" gorm:"type:text"`
	CreatedAt   time.Time        `json:"created_at"`
	DeletedAt   gorm.DeletedAt   `json:"-" gorm:"index"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
