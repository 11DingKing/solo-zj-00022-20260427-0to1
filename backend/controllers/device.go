package controllers

import (
	"net/http"
	"repair-system/models"
	"repair-system/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateDeviceRequest struct {
	DeviceCode         string `json:"device_code" binding:"required"`
	Name               string `json:"name" binding:"required"`
	Model              string `json:"model"`
	Location           string `json:"location"`
	PurchaseDate       string `json:"purchase_date"`
	WarrantyExpireDate string `json:"warranty_expire_date"`
}

type UpdateDeviceRequest struct {
	Name               string `json:"name"`
	Model              string `json:"model"`
	Location           string `json:"location"`
	PurchaseDate       string `json:"purchase_date"`
	WarrantyExpireDate string `json:"warranty_expire_date"`
	Status             string `json:"status"`
}

func GetDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	query := utils.DB.Model(&models.Device{})

	if keyword != "" {
		query = query.Where("device_code LIKE ? OR name LIKE ? OR model LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	var devices []models.Device
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      devices,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetDeviceByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	var device models.Device
	if err := utils.DB.First(&device, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	var workOrders []models.WorkOrder
	if err := utils.DB.Where("device_id = ?", id).
		Preload("Employee").
		Preload("Technician").
		Order("created_at DESC").
		Find(&workOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch work orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device":      device,
		"work_orders": workOrders,
	})
}

func CreateDevice(c *gin.Context) {
	var req CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var existingDevice models.Device
	if err := utils.DB.Where("device_code = ?", req.DeviceCode).First(&existingDevice).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Device code already exists"})
		return
	}

	device := models.Device{
		DeviceCode: req.DeviceCode,
		Name:       req.Name,
		Model:      req.Model,
		Location:   req.Location,
		Status:     models.DeviceStatusActive,
	}

	if req.PurchaseDate != "" {
		t, err := time.Parse("2006-01-02", req.PurchaseDate)
		if err == nil {
			device.PurchaseDate = &t
		}
	}

	if req.WarrantyExpireDate != "" {
		t, err := time.Parse("2006-01-02", req.WarrantyExpireDate)
		if err == nil {
			device.WarrantyExpireDate = &t
		}
	}

	if err := utils.DB.Create(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create device"})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func UpdateDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	var device models.Device
	if err := utils.DB.First(&device, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Name != "" {
		device.Name = req.Name
	}
	if req.Model != "" {
		device.Model = req.Model
	}
	if req.Location != "" {
		device.Location = req.Location
	}
	if req.Status != "" {
		device.Status = models.DeviceStatus(req.Status)
	}

	if req.PurchaseDate != "" {
		t, err := time.Parse("2006-01-02", req.PurchaseDate)
		if err == nil {
			device.PurchaseDate = &t
		}
	}

	if req.WarrantyExpireDate != "" {
		t, err := time.Parse("2006-01-02", req.WarrantyExpireDate)
		if err == nil {
			device.WarrantyExpireDate = &t
		}
	}

	if err := utils.DB.Save(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func DeleteDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	var count int64
	utils.DB.Model(&models.WorkOrder{}).Where("device_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete device with existing work orders"})
		return
	}

	var device models.Device
	if err := utils.DB.First(&device, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	if err := utils.DB.Delete(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}
