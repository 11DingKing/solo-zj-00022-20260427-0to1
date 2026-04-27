package controllers

import (
	"net/http"
	"repair-system/models"
	"repair-system/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Username string          `json:"username" binding:"required"`
	Password string          `json:"password" binding:"required"`
	RealName string          `json:"real_name" binding:"required"`
	Role     models.UserRole `json:"role" binding:"required"`
	Phone    string          `json:"phone"`
}

type UpdateUserRequest struct {
	RealName string          `json:"real_name"`
	Role     models.UserRole `json:"role"`
	Phone    string          `json:"phone"`
	Password string          `json:"password"`
	Status   models.UserStatus `json:"status"`
}

func GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	role := c.Query("role")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	query := utils.DB.Model(&models.User{})

	if role != "" {
		query = query.Where("role = ?", role)
	}
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	var users []models.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := utils.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var existingUser models.User
	if err := utils.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	user := models.User{
		Username: req.Username,
		Password: req.Password,
		RealName: req.RealName,
		Role:     req.Role,
		Phone:    req.Phone,
	}

	if err := utils.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := utils.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Password != "" {
		user.Password = req.Password
		if err := user.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := utils.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	currentUserID, _ := c.Get("user_id")
	if uint(id) == currentUserID.(uint) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete yourself"})
		return
	}

	var user models.User
	if err := utils.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := utils.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func GetTechnicians(c *gin.Context) {
	var technicians []models.User
	if err := utils.DB.Where("role = ?", models.RoleTechnician).Order("created_at DESC").Find(&technicians).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch technicians"})
		return
	}

	c.JSON(http.StatusOK, technicians)
}

func ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := utils.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Password = "123456"
	if err := user.HashPassword(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := utils.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
