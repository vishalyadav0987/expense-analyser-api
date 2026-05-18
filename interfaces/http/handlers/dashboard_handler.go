package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vishalyadav0987/expense-analyser/internal/application/dashboard"
)

type DashboardHandler struct {
	service dashboard.DashboardService
}

func NewDashboardHandler(service dashboard.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) HandleGetSummary(c *gin.Context) {
	userID := c.GetString("userID")

	// Get Month/Year from Query params, default to current Time
	now := time.Now()
	monthStr := c.DefaultQuery("month", fmt.Sprintf("%d", int(now.Month())))
	yearStr := c.DefaultQuery("year", fmt.Sprintf("%d", now.Year()))

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	data, err := h.service.GetDashboardSummary(c.Request.Context(), userID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dashboard data fetched successfully.",
		"data":    data,
	})
}
