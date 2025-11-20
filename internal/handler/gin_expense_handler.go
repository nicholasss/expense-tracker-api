package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
)

// === Package Variables

var (
	ErrMissingHeader = fmt.Errorf("header is missing")
	ErrInvalidID     = fmt.Errorf("provided id is invalid")
)

// === Helper Functions

// ginValidateID will take an id of string type and return the int type or an error.
// It will also validate that it can be used with the service layer to get a record.
func ginValidateID(idStr string) (int, error) {
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}

	if idInt <= 0 {
		return 0, ErrInvalidID
	}

	return idInt, nil
}

// === Handler

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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	responseRecords := make([]*ExpenseResponse, 0)
	for _, record := range records {
		responseRecords = append(responseRecords, expenseToResponse(record))
	}

	// send data
	c.JSON(http.StatusOK, responseRecords)
}

func (h *GinHandler) GetExpenseByID(c *gin.Context) {
	// check the ID for validity
	idInt, err := ginValidateID(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request: " + err.Error()})
		return
	}

	// get the record
	record, err := h.Service.GetExpenseByID(c.Request.Context(), idInt)
	if err != nil {
		log.Printf("error: %s", err)
		// specifically respond 404 if id is not a record
		if errors.Is(err, expenses.ErrUnusedID) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not Found: " + err.Error()})
			return
		}

		// otherwise send generic error
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// send reccord
	c.JSON(http.StatusOK, expenseToResponse(record))
}
