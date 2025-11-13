// Package handler utilizes the controller (handler) pattern in order to handle the web request logic.
//
// One of its responsibilities is to validate requests and route them to the relevant services.
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"
)

// === Types ===

// == Handler Definition ==

// The ExpenseHandler is imported into a /cmd package, and provides the handlers via its methods.
type ExpenseHandler struct {
	Service *expenses.Service
}

// NewExpanseHandler does not use an interface, but rather directly using the exported Service type.
func NewExpanseHandler(service *expenses.Service) *ExpenseHandler {
	return &ExpenseHandler{Service: service}
}

// == Helper Types ==

// RFC3339Time is a type that wraps and implements time.Time as a un/marshal-able type
type RFC3339Time struct {
	time.Time
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
type CreateExpenseRequest struct {
	OccuredAt   RFC3339Time `json:"occured_at"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
}

// validate performs structural/syntactic validation
//
// Method validate will respond with a slice of string, and a bool.
// If it is valid then it will be `[]string{}, true`,
// otherwise it will list reasons and be false `[]string{...}, false`.
// You could utilize the 'comma ok' idiom or not, as shown below.
//
// reasons, isValid := c.validate()
func (c *CreateExpenseRequest) validate() ([]string, bool) {
	issues := make([]string, 0)
	isValid := true

	if c.Amount <= 0 {
		issues = append(issues, "field 'amount' is negative, missing, or zero")
		isValid = false
	}
	if c.OccuredAt.IsZero() {
		issues = append(issues, "field 'occured_at' is missing or empty")
		isValid = false
	}
	if c.Description == "" {
		issues = append(issues, "field 'description' is missing or empty")
		isValid = false
	}
	return issues, isValid
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

// === Helper Functions ===

// headersAreValid will check for missing headers and will call sendError if needed.
func (h *ExpenseHandler) headersAreValid(w http.ResponseWriter, r *http.Request) bool {
	issues := make([]string, 0)

	ct := r.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		issues = append(issues, "'Content-Type' header missing 'application/json'")
	}

	if len(issues) == 0 {
		return true
	}

	h.sendErrors(w, 400, issues)
	return false
}

// sendJSON handles errors internaly and will write directly to the client where able.
func (h *ExpenseHandler) sendJSON(w http.ResponseWriter, status int, responsePayload any) {
	// marshal response payload
	respData, err := json.Marshal(responsePayload)
	if err != nil {
		h.sendErrors(w, 500, []string{"database error"})
		return
	}

	// first set header and response code
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	bytesWritten, err := w.Write(respData)
	if err != nil {
		log.Println("unable to write full response due to:", err)
		return
	}
	if bytesWritten < len(respData) {
		log.Println("unable to write full response based on bytes written")
		return
	}
}

// sendErrors will create and write the provided error code to the provided http.ResponseWriter.
func (h *ExpenseHandler) sendErrors(w http.ResponseWriter, code int, issues []string) {
	if len(issues) == 0 {
		issues = append(issues, http.StatusText(code))
	}

	payload := &ErrorResponse{
		HTTPCode: code,
		Issues:   issues,
	}
	issuesData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	w.WriteHeader(code)
	bytesWritten, err := w.Write(issuesData)
	if err != nil {
		log.Println("unable to write full response due to:", err)
		return
	}
	if bytesWritten < len(issuesData) {
		log.Println("unable to write full response based on bytes written")
		return
	}
}

// === Endpoint Handlers ===

// GetAllExpenses ...
func (h *ExpenseHandler) GetAllExpenses(w http.ResponseWriter, r *http.Request) {
	expRecords, err := h.Service.GetAllExpenses(r.Context())
	if err != nil {
		h.sendErrors(w, 500, []string{"database error"})
		return
	}

	responsePayload := make([]ExpenseResponse, 0, len(expRecords))
	for _, exp := range expRecords {
		responsePayload = append(responsePayload, *expenseToResponse(exp))
	}

	h.sendJSON(w, 200, responsePayload)
}

// CreateExpense handles 'POST /expenses'
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	if !h.headersAreValid(w, r) {
		return
	}

	// get the json body, partially performs validation by ensuring structural validity
	var reqBody *CreateExpenseRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		h.sendErrors(w, 400, []string{"unable to decode body"})
		return
	}

	// defer body closure and error check
	defer func() {
		err = r.Body.Close()
		if err != nil {
			h.sendErrors(w, 500, []string{})
		}
	}()

	// validation of structure
	issues, isValid := reqBody.validate()
	if !isValid {
		h.sendErrors(w, 400, issues)
		return
	}

	// send to service layer and semantic/'business' validation
	//
	// we will just need to send the single error back to the client, even if there are multiple issues
	expRecord, err := h.Service.NewExpense(r.Context(), reqBody.OccuredAt.Time, reqBody.Description, reqBody.Amount)
	if err != nil {
		// check for custom errors
		if errors.Is(err, expenses.ErrInvalidOccuredAtTime) {
			h.sendErrors(w, http.StatusBadRequest, []string{err.Error()})
			return
		}
		if errors.Is(err, expenses.ErrInvalidDescription) {
			h.sendErrors(w, http.StatusBadRequest, []string{err.Error()})
			return
		}
		if errors.Is(err, expenses.ErrInvalidAmount) {
			h.sendErrors(w, http.StatusBadRequest, []string{err.Error()})
			return
		}

		// otherwise errors would likely be 5XX
		h.sendErrors(w, 500, []string{})
		return
	}

	responsePayload := expenseToResponse(expRecord)
	h.sendJSON(w, 201, responsePayload)
}

// GetExpenseByID ...
func (h *ExpenseHandler) GetExpenseByID(w http.ResponseWriter, r *http.Request) {
	log.Println("get expense by id not implemented yet")
}

// UpdateExpense ...
func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	log.Println("update expense is not implemented yet")
}

// DeleteExpense ...
func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	log.Println("delete expense is not implemented yet")
}

// SummarizeExpenses ...
func (h *ExpenseHandler) SummarizeExpenses(w http.ResponseWriter, r *http.Request) {
	log.Println("summarize expenses is not implmeneted yet")
}
