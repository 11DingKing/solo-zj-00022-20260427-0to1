package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type WorkOrderStatus string

const (
	StatusPendingAssign   WorkOrderStatus = "pending_assign"
	StatusAssigned        WorkOrderStatus = "assigned"
	StatusProcessing      WorkOrderStatus = "processing"
	StatusPendingConfirm  WorkOrderStatus = "pending_confirm"
	StatusClosed          WorkOrderStatus = "closed"
)

type FaultType string

const (
	FaultTypeHardware FaultType = "hardware"
	FaultTypeSoftware FaultType = "software"
	FaultTypeNetwork  FaultType = "network"
	FaultTypeOther    FaultType = "other"
)

type UrgencyLevel string

const (
	UrgencyLow     UrgencyLevel = "low"
	UrgencyMedium  UrgencyLevel = "medium"
	UrgencyHigh    UrgencyLevel = "high"
	UrgencyUrgent  UrgencyLevel = "urgent"
)

type WorkOrder struct {
	ID              uint             `json:"id" gorm:"primaryKey"`
	OrderNumber     string           `json:"order_number" gorm:"uniqueIndex;size:50;not null"`
	DeviceID        uint             `json:"device_id" gorm:"not null"`
	EmployeeID      uint             `json:"employee_id" gorm:"not null"`
	TechnicianID    *uint            `json:"technician_id"`
	FaultType       FaultType        `json:"fault_type" gorm:"type:enum('hardware','software','network','other');not null"`
	FaultDescription string          `json:"fault_description" gorm:"type:text;not null"`
	Urgency         UrgencyLevel     `json:"urgency" gorm:"type:enum('low','medium','high','urgent');default:'low';not null"`
	Status          WorkOrderStatus  `json:"status" gorm:"type:enum('pending_assign','assigned','processing','pending_confirm','closed');default:'pending_assign';not null"`
	RepairMeasures  string           `json:"repair_measures" gorm:"type:text"`
	ReplacedParts   string           `json:"replaced_parts" gorm:"type:text"`
	RepairDuration  int              `json:"repair_duration" gorm:"comment:'维修耗时，单位：分钟'"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `json:"-" gorm:"index"`

	Device      *Device  `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
	Employee    *User    `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
	Technician  *User    `json:"technician,omitempty" gorm:"foreignKey:TechnicianID"`
	Images      []Image  `json:"images,omitempty" gorm:"foreignKey:WorkOrderID"`
	Logs        []OperationLog `json:"logs,omitempty" gorm:"foreignKey:WorkOrderID"`
}

func (wo *WorkOrder) BeforeCreate(tx *gorm.DB) error {
	if wo.OrderNumber == "" {
		wo.OrderNumber = GenerateOrderNumber()
	}
	return nil
}

func GenerateOrderNumber() string {
	now := time.Now()
	return fmt.Sprintf("WO%s%06d", now.Format("20060102"), now.UnixNano()%1000000)
}

var StatusTransitions = map[WorkOrderStatus][]WorkOrderStatus{
	StatusPendingAssign:  {StatusAssigned},
	StatusAssigned:       {StatusProcessing},
	StatusProcessing:     {StatusPendingConfirm},
	StatusPendingConfirm: {StatusClosed, StatusProcessing},
	StatusClosed:         {},
}

func (wo *WorkOrder) CanTransitionTo(newStatus WorkOrderStatus) bool {
	allowedStatuses, exists := StatusTransitions[wo.Status]
	if !exists {
		return false
	}
	for _, s := range allowedStatuses {
		if s == newStatus {
			return true
		}
	}
	return false
}
