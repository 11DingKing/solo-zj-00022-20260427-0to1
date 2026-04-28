package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"repair-system/config"
	"repair-system/models"
	"repair-system/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only jpg, jpeg, png, gif, webp are allowed"})
		return
	}

	maxSize := int64(10 * 1024 * 1024)
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10MB limit"})
		return
	}

	uploadDir := config.AppConfig.UploadDir
	dateDir := time.Now().Format("2006/01/02")
	fullDir := filepath.Join(uploadDir, dateDir)

	if err := utils.EnsureDir(fullDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(fullDir, fileName)
	relativePath := fmt.Sprintf("%s/%s", dateDir, fileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	image := models.Image{
		FilePath:  relativePath,
		FileName:  file.Filename,
		FileSize:  int(file.Size),
		ImageType: models.ImageTypeBefore,
	}

	if err := utils.DB.Create(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         image.ID,
		"file_path":  relativePath,
		"file_name":  file.Filename,
		"file_size":  file.Size,
		"access_url": fmt.Sprintf("/uploads/%s", relativePath),
	})
}
