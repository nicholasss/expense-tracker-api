package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
)

var ErrMissingHeader = fmt.Errorf("header is missing")

type GinHandler struct {
	Service expenses.Service
}

func NewGinHandler(service expenses.Service) *GinHandler {
	return &GinHandler{Service: service}
}

func (h *GinHandler) GetAllExpenses(c *gin.Context) {
	// contentType := c.GetHeader("Content-Type")
	// if contentType != "application/json" {
	// 	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrMissingHeader})
	// 	return
	// }

	// get data
	records, err := h.Service.GetAllExpenses(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	responseRecords := make([]*ExpenseResponse, 0)
	for _, record := range records {
		responseRecords = append(responseRecords, expenseToResponse(record))
	}

	// send data
	c.JSON(http.StatusOK, responseRecords)
}
