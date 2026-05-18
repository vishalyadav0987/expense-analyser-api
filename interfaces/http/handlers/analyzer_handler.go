package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vishalyadav0987/expense-analyser/interfaces/http/dto"
	"github.com/vishalyadav0987/expense-analyser/internal/application/analyzer"
)

type AnalyzerHandler struct {
	service *analyzer.AnalyzerService
}

func NewAnalyzerHandler(service *analyzer.AnalyzerService) *AnalyzerHandler {
	return &AnalyzerHandler{service: service}
}

func (h *AnalyzerHandler) HandleGetWeekly(c *gin.Context) {
	userID, _ := c.Get("userID")

	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("startDate and endDate are required parameters"))
		return
	}

	// Parse dates (Expects YYYY-MM-DD)
	// SDE3 Catch: Ensure EndDate covers the ENTIRE day till 23:59:59
	startDate, err1 := time.Parse("2006-01-02", startDateStr)
	endDate, err2 := time.Parse("2006-01-02", endDateStr)

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid date format, use YYYY-MM-DD"))
		return
	}

	// Make end date the end of the day to catch evening expenses
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	report, err := h.service.GenerateWeeklyReport(c.Request.Context(), userID.(string), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("failed to generate report"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}
