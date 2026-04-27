package controllers

import (
	"net/http"
	"repair-system/models"
	"repair-system/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateWorkOrderRequest struct {
	DeviceID         uint                `json:"device_id" binding:"required"`
	FaultType        models.FaultType    `json:"fault_type" binding:"required"`
	FaultDescription string              `json:"fault_description" binding:"required"`
	Urgency          models.UrgencyLevel `json:"urgency"`
	BeforeImageIDs   []uint              `json:"before_image_ids"`
}

type AssignWorkOrderRequest struct {
	TechnicianID uint `json:"technician_id" binding:"required"`
}

type SubmitRepairRequest struct {
	RepairMeasures string `json:"repair_measures" binding:"required"`
	ReplacedParts  string `json:"replaced_parts"`
	RepairDuration int    `json:"repair_duration"`
	AfterImageIDs  []uint `json:"after_image_ids"`
}

func GetWorkOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	urgency := c.Query("urgency")
	faultType := c.Query("fault_type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	query := utils.DB.Model(&models.WorkOrder{})

	switch userRole.(string) {
	case string(models.RoleEmployee):
		query = query.Where("employee_id = ?", userID)
	case string(models.RoleTechnician):
		query = query.Where("technician_id = ?", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if urgency != "" {
		query = query.Where("urgency = ?", urgency)
	}
	if faultType != "" {
		query = query.Where("fault_type = ?", faultType)
	}
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	query.Count(&total)

	var workOrders []models.WorkOrder
	offset := (page - 1) * pageSize
	if err := query.Preload("Device").
		Preload("Employee").
		Preload("Technician").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&workOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch work orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      workOrders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetWorkOrderByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work order ID"})
		return
	}

	var workOrder models.WorkOrder
	if err := utils.DB.Preload("Device").
		Preload("Employee").
		Preload("Technician").
		Preload("Images").
		Preload("Logs.User").
		First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work order not found"})
		return
	}

	c.JSON(http.StatusOK, workOrder)
}

func CreateWorkOrder(c *gin.Context) {
	var req CreateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID, _ := c.Get("user_id")

	var device models.Device
	if err := utils.DB.First(&device, req.DeviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	workOrder := models.WorkOrder{
		DeviceID:         req.DeviceID,
		EmployeeID:       userID.(uint),
		FaultType:        req.FaultType,
		FaultDescription: req.FaultDescription,
		Urgency:          req.Urgency,
		Status:           models.StatusPendingAssign,
	}

	if workOrder.Urgency == "" {
		workOrder.Urgency = models.UrgencyLow
	}

	tx := utils.DB.Begin()

	if err := tx.Create(&workOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create work order"})
		return
	}

	if len(req.BeforeImageIDs) > 0 && len(req.BeforeImageIDs) <= 4 {
		for _, imgID := range req.BeforeImageIDs {
			var image models.Image
			if err := tx.First(&image, imgID).Error; err == nil {
				image.WorkOrderID = workOrder.ID
				image.ImageType = models.ImageTypeBefore
				tx.Save(&image)
			}
		}
	}

	log := models.OperationLog{
		WorkOrderID: workOrder.ID,
		UserID:      userID.(uint),
		Operation:   models.OperationCreate,
		NewStatus:   &workOrder.Status,
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation log"})
		return
	}

	tx.Commit()

	if err := utils.DB.Preload("Device").
		Preload("Employee").
		First(&workOrder, workOrder.ID).Error; err != nil {
		c.JSON(http.StatusCreated, workOrder)
		return
	}

	c.JSON(http.StatusCreated, workOrder)
}

func AssignWorkOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work order ID"})
		return
	}

	var req AssignWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID, _ := c.Get("user_id")

	var workOrder models.WorkOrder
	if err := utils.DB.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work order not found"})
		return
	}

	if workOrder.Status != models.StatusPendingAssign {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Work order is not in pending_assign status"})
		return
	}

	var technician models.User
	if err := utils.DB.Where("id = ? AND role = ?", req.TechnicianID, models.RoleTechnician).First(&technician).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Technician not found"})
		return
	}

	oldStatus := workOrder.Status
	newStatus := models.StatusAssigned

	tx := utils.DB.Begin()

	workOrder.TechnicianID = &req.TechnicianID
	workOrder.Status = newStatus
	if err := tx.Save(&workOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign work order"})
		return
	}

	log := models.OperationLog{
		WorkOrderID: workOrder.ID,
		UserID:      userID.(uint),
		Operation:   models.OperationAssign,
		OldStatus:   &oldStatus,
		NewStatus:   &newStatus,
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation log"})
		return
	}

	tx.Commit()

	if err := utils.DB.Preload("Device").
		Preload("Employee").
		Preload("Technician").
		First(&workOrder, workOrder.ID).Error; err != nil {
		c.JSON(http.StatusOK, workOrder)
		return
	}

	c.JSON(http.StatusOK, workOrder)
}

func AcceptWorkOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work order ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var workOrder models.WorkOrder
	if err := utils.DB.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work order not found"})
		return
	}

	if workOrder.Status != models.StatusAssigned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Work order is not in assigned status"})
		return
	}

	if workOrder.TechnicianID == nil || *workOrder.TechnicianID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not assigned to this work order"})
		return
	}

	oldStatus := workOrder.Status
	newStatus := models.StatusProcessing

	tx := utils.DB.Begin()

	workOrder.Status = newStatus
	if err := tx.Save(&workOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept work order"})
		return
	}

	log := models.OperationLog{
		WorkOrderID: workOrder.ID,
		UserID:      userID.(uint),
		Operation:   models.OperationAccept,
		OldStatus:   &oldStatus,
		NewStatus:   &newStatus,
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation log"})
		return
	}

	tx.Commit()

	if err := utils.DB.Preload("Device").
		Preload("Employee").
		Preload("Technician").
		First(&workOrder, workOrder.ID).Error; err != nil {
		c.JSON(http.StatusOK, workOrder)
		return
	}

	c.JSON(http.StatusOK, workOrder)
}

func SubmitRepair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work order ID"})
		return
	}

	var req SubmitRepairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID, _ := c.Get("user_id")

	var workOrder models.WorkOrder
	if err := utils.DB.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work order not found"})
		return
	}

	if workOrder.Status != models.StatusProcessing {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Work order is not in processing status"})
		return
	}

	if workOrder.TechnicianID == nil || *workOrder.TechnicianID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not assigned to this work order"})
		return
	}

	oldStatus := workOrder.Status
	newStatus := models.StatusPendingConfirm

	tx := utils.DB.Begin()

	workOrder.RepairMeasures = req.RepairMeasures
	workOrder.ReplacedParts = req.ReplacedParts
	workOrder.RepairDuration = req.RepairDuration
	workOrder.Status = newStatus
	if err := tx.Save(&workOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit repair"})
		return
	}

	if len(req.AfterImageIDs) > 0 && len(req.AfterImageIDs) <= 4 {
		for _, imgID := range req.AfterImageIDs {
			var image models.Image
			if err := tx.First(&image, imgID).Error; err == nil {
				image.WorkOrderID = workOrder.ID
				image.ImageType = models.ImageTypeAfter
				tx.Save(&image)
			}
		}
	}

	log := models.OperationLog{
		WorkOrderID: workOrder.ID,
		UserID:      userID.(uint),
		Operation:   models.OperationSubmit,
		OldStatus:   &oldStatus,
		NewStatus:   &newStatus,
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation log"})
		return
	}

	tx.Commit()

	if err := utils.DB.Preload("Device").
		Preload("Employee").
		Preload("Technician").
		First(&workOrder, workOrder.ID).Error; err != nil {
		c.JSON(http.StatusOK, workOrder)
		return
	}

	c.JSON(http.StatusOK, workOrder)
}

func ConfirmWorkOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work order ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var workOrder models.WorkOrder
	if err := utils.DB.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work order not found"})
		return
	}

	if workOrder.Status != models.StatusPendingConfirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Work order is not in pending_confirm status"})
		return
	}

	if workOrder.EmployeeID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of this work order"})
		return
	}

	oldStatus := workOrder.Status
	newStatus := models.StatusClosed

	tx := utils.DB.Begin()

	workOrder.Status = newStatus
	if err := tx.Save(&workOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm work order"})
		return
	}

	log := models.OperationLog{
		WorkOrderID: workOrder.ID,
		UserID:      userID.(uint),
		Operation:   models.OperationConfirm,
		OldStatus:   &oldStatus,
		NewStatus:   &newStatus,
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation log"})
		return
	}

	tx.Commit()

	if err := utils.DB.Preload("Device").
		Preload("Employee").
		Preload("Technician").
		First(&workOrder, workOrder.ID).Error; err != nil {
		c.JSON(http.StatusOK, workOrder)
		return
	}

	c.JSON(http.StatusOK, workOrder)
}

func RejectWorkOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work order ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var workOrder models.WorkOrder
	if err := utils.DB.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work order not found"})
		return
	}

	if workOrder.Status != models.StatusPendingConfirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Work order is not in pending_confirm status"})
		return
	}

	if workOrder.EmployeeID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of this work order"})
		return
	}

	oldStatus := workOrder.Status
	newStatus := models.StatusProcessing

	tx := utils.DB.Begin()

	workOrder.Status = newStatus
	if err := tx.Save(&workOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject work order"})
		return
	}

	log := models.OperationLog{
		WorkOrderID: workOrder.ID,
		UserID:      userID.(uint),
		Operation:   models.OperationReject,
		OldStatus:   &oldStatus,
		NewStatus:   &newStatus,
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation log"})
		return
	}

	tx.Commit()

	if err := utils.DB.Preload("Device").
		Preload("Employee").
		Preload("Technician").
		First(&workOrder, workOrder.ID).Error; err != nil {
		c.JSON(http.StatusOK, workOrder)
		return
	}

	c.JSON(http.StatusOK, workOrder)
}
