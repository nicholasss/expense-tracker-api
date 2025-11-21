package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
)

// === Package Variables

var (
	ErrMissingHeader = fmt.Errorf("one or more header(s) are missing")
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

func (h *GinHandler) CreateExpense(c *gin.Context) {
	// request body bind
	var reqBody CreateExpenseRequest
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request: " + err.Error()})
		return
	}

	// send to service layer
	newRecord, err := h.Service.NewExpense(c.Request.Context(), reqBody.OccuredAt.Time, reqBody.Description, reqBody.Amount)
	if err != nil {
		// checking for service errors
		if errors.Is(err, expenses.ErrInvalidAmount) || errors.Is(err, expenses.ErrInvalidOccuredAtTime) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request: " + err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// return record
	c.JSON(http.StatusCreated, expenseToResponse(newRecord))
}

func (h *GinHandler) UpdateExpense(c *gin.Context) {
	// bind and validation
	var reqBody UpdateExpenseRequest
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request: " + err.Error()})
		return
	}

	// send to service layer
	err = h.Service.UpdateExpense(c.Request.Context(), reqBody.ID, reqBody.OccuredAt.Time, reqBody.Description, reqBody.Amount)
	if err != nil {
		if errors.Is(err, expenses.ErrInvalidAmount) || errors.Is(err, expenses.ErrInvalidOccuredAtTime) {
			// service error
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request: " + err.Error()})
			return
		} else if errors.Is(err, expenses.ErrUnusedID) {
			// repository error
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		// generic error
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// all went well
	c.Status(http.StatusNoContent)
}

func (h *GinHandler) DeleteExpense(c *gin.Context) {
	// check the ID for validity
	idInt, err := ginValidateID(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request: " + err.Error()})
		return
	}

	// delete the record
	err = h.Service.DeleteExpense(c.Request.Context(), idInt)
	if err != nil {
		// repository errors
		if errors.Is(err, expenses.ErrInvalidID) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request: " + err.Error()})
			return
		} else if errors.Is(err, expenses.ErrUnusedID) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		// generic server error
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.Status(http.StatusNoContent)
}
