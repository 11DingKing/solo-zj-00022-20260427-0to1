package controllers

import (
	"net/http"
	"repair-system/models"
	"repair-system/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardStats struct {
	TodayNewOrders     int64   `json:"today_new_orders"`
	PendingOrders      int64   `json:"pending_orders"`
	AvgProcessingTime  float64 `json:"avg_processing_time"`
}

type FaultTypeDistribution struct {
	FaultType string `json:"fault_type"`
	Count     int64  `json:"count"`
}

type DailyTrend struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type TechnicianRanking struct {
	TechnicianID   uint   `json:"technician_id"`
	TechnicianName string `json:"technician_name"`
	CompletedCount int64  `json:"completed_count"`
}

func GetDashboardStats(c *gin.Context) {
	var stats DashboardStats

	today := time.Now().Format("2006-01-02")
	utils.DB.Model(&models.WorkOrder{}).Where("DATE(created_at) = ?", today).Count(&stats.TodayNewOrders)

	utils.DB.Model(&models.WorkOrder{}).Where("status IN ?", []string{
		string(models.StatusPendingAssign),
		string(models.StatusAssigned),
		string(models.StatusProcessing),
	}).Count(&stats.PendingOrders)

	var totalDuration int64
	var closedCount int64
	type Result struct {
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	var results []Result

	utils.DB.Model(&models.WorkOrder{}).
		Where("status = ? AND repair_duration IS NOT NULL", models.StatusClosed).
		Select("created_at", "updated_at").
		Find(&results)

	closedCount = int64(len(results))
	if closedCount > 0 {
		for _, r := range results {
			duration := r.UpdatedAt.Sub(r.CreatedAt).Minutes()
			totalDuration += int64(duration)
		}
		stats.AvgProcessingTime = float64(totalDuration) / float64(closedCount)
	}

	c.JSON(http.StatusOK, stats)
}

func GetFaultTypeDistribution(c *gin.Context) {
	var distributions []FaultTypeDistribution

	utils.DB.Model(&models.WorkOrder{}).
		Select("fault_type, COUNT(*) as count").
		Group("fault_type").
		Scan(&distributions)

	c.JSON(http.StatusOK, distributions)
}

func GetLast30DaysTrend(c *gin.Context) {
	var trends []DailyTrend

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	utils.DB.Model(&models.WorkOrder{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("DATE(created_at) >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&trends)

	c.JSON(http.StatusOK, trends)
}

func GetTechnicianRanking(c *gin.Context) {
	var rankings []TechnicianRanking

	utils.DB.Model(&models.WorkOrder{}).
		Select("work_orders.technician_id, users.real_name as technician_name, COUNT(*) as completed_count").
		Joins("LEFT JOIN users ON work_orders.technician_id = users.id").
		Where("work_orders.status = ?", models.StatusClosed).
		Where("work_orders.technician_id IS NOT NULL").
		Group("work_orders.technician_id, users.real_name").
		Order("completed_count DESC").
		Limit(10).
		Scan(&rankings)

	c.JSON(http.StatusOK, rankings)
}
