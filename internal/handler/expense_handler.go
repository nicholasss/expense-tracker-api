package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
)

// === Handler Type

type GinHandler struct {
	Service expenses.Service
}

func NewGinHandler(service expenses.Service) *GinHandler {
	return &GinHandler{Service: service}
}

// == Helper Types ==

// RFC3339Time is a type that wraps and implements time.Time as a un/marshal-able type
// NOTE: because this is a struct itself, `validator` expects the field here, not on the request struct.
type RFC3339Time struct {
	time.Time `binding:"required"`
}

func (t *RFC3339Time) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	t.Time = parsed
	return nil
}

func (t *RFC3339Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(time.RFC3339))
}

// == Endpoint Types ==

// CreateExpenseRequest is utilized specifically for the CreateExpense endpoint: POST /expense
// NOTE: While `validator` can perfrom recursive checking of binding:"", it seems to only do that for struct types.
type CreateExpenseRequest struct {
	OccuredAt   RFC3339Time `json:"occured_at"`
	Description string      `json:"description" binding:"required"`
	Amount      int64       `json:"amount" binding:"required,gt=0"`
}

// UpdateExpenseRequest is utilized specifically for the UpdateExpense endpoint: PUT /expense
type UpdateExpenseRequest struct {
	ID int `json:"id" binding:"required"`
	CreateExpenseRequest
}

// ExpenseResponse is hopefully a general response that can be used across several endpoints
type ExpenseResponse struct {
	ID          int         `json:"id"`
	CreatedAt   RFC3339Time `json:"created_at"`
	OccuredAt   RFC3339Time `json:"occured_at"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
}

func expenseToResponse(exp *expenses.Expense) *ExpenseResponse {
	return &ExpenseResponse{
		ID:          exp.ID,
		CreatedAt:   RFC3339Time{Time: exp.RecordCreatedAt},
		OccuredAt:   RFC3339Time{Time: exp.ExpenseOccuredAt},
		Description: exp.Description,
		Amount:      exp.Amount,
	}
}

// ErrorResponse is a payload type that is used for sending errors to the clients.
type ErrorResponse struct {
	HTTPCode int      `json:"code"`
	Issues   []string `json:"issues"`
}

// === Endpoint Hanlders ===

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
	idInt, err := strconv.Atoi(c.Param("id"))
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
	idInt, err := strconv.Atoi(c.Param("id"))
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
