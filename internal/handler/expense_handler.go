// Package handler utilizes the controller (handler) pattern in order to handle the web request logic.
//
// One of its responsibilities is to validate requests and route them to the relevant services.
package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"
)

// === Types ===

// == Handler Definition ==

// The ExpenseHandler is imported into a /cmd package, and provides the handlers via its methods.
type ExpenseHandler struct {
	Service expenses.Service
}

func NewExpanseHandler(service expenses.Service) *ExpenseHandler {
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

// UpdateExpenseRequest is utilized specifically for the UpdateExpense endpoint: PUT /expense
type UpdateExpenseRequest struct {
	ID int `json:"id"`
	CreateExpenseRequest
}

func (u *UpdateExpenseRequest) validate() ([]string, bool) {
	issues := make([]string, 0)
	isValid := true

	if u.ID <= 0 {
		issues = append(issues, "id is not valid")
		isValid = false
	}

	// just utilize the prexisting validation method
	c := CreateExpenseRequest{Amount: u.Amount, Description: u.Description, OccuredAt: u.OccuredAt}
	cIssues, cIsValid := c.validate()
	if !cIsValid {
		issues = append(issues, cIssues...)
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

// validateID performs structural validation of the ID
func validateID(idStr string) (int, error) {
	if idStr == "" {
		return 0, errors.New("id in path is missing")
	}

	// checking structural validity
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("id in path is not valid id")
	}

	return idInt, nil
}

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

	h.sendErrors(w, http.StatusBadRequest, issues)
	return false
}

// sendJSON handles errors internaly and will write directly to the client where able.
func (h *ExpenseHandler) sendJSON(w http.ResponseWriter, status int, responsePayload any) {
	// marshal response payload
	respData, err := json.Marshal(responsePayload)
	if err != nil {
		h.sendErrors(w, http.StatusInternalServerError, []string{"database error"})
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
	// guard against invalid status codes
	if statusText := http.StatusText(code); statusText == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// guard against nil or empty slices
	if issues == nil {
		issues = make([]string, 0)
	}
	if len(issues) == 0 {
		issues = append(issues, http.StatusText(code))
	}

	// marshal to bytes
	payload := &ErrorResponse{
		HTTPCode: code,
		Issues:   issues,
	}
	issuesData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// send to client
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
		h.sendErrors(w, http.StatusInternalServerError, []string{"database error"})
		return
	}

	responsePayload := make([]ExpenseResponse, 0, len(expRecords))
	for _, exp := range expRecords {
		responsePayload = append(responsePayload, *expenseToResponse(exp))
	}

	h.sendJSON(w, http.StatusOK, responsePayload)
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
		h.sendErrors(w, http.StatusBadRequest, []string{"unable to decode body"})
		return
	}

	// defer body closure and error check
	defer func() {
		err = r.Body.Close()
		if err != nil {
			h.sendErrors(w, http.StatusInternalServerError, []string{})
		}
	}()

	// validation of structure
	issues, isValid := reqBody.validate()
	if !isValid {
		h.sendErrors(w, http.StatusBadRequest, issues)
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
	// structural validation of id
	idInt, err := validateID(r.PathValue("id"))
	if err != nil {
		h.sendErrors(w, http.StatusBadRequest, []string{err.Error()})
		return
	}

	// calling service and letting it perform semantic/'business' validation
	exp, err := h.Service.GetExpenseByID(r.Context(), idInt)
	if err != nil {
		// checking for invalid ID or an ID that does not exist
		if errors.Is(err, expenses.ErrInvalidID) {
			h.sendErrors(w, http.StatusBadRequest, []string{err.Error()})
			return
		}
		if errors.Is(err, expenses.ErrUnusedID) || errors.Is(err, sql.ErrNoRows) {
			h.sendErrors(w, http.StatusNotFound, []string{err.Error()})
			return
		}
		// any other errors
		h.sendErrors(w, http.StatusInternalServerError, []string{})
		return
	}

	responsePayload := expenseToResponse(exp)
	h.sendJSON(w, http.StatusOK, responsePayload)
}

// UpdateExpense ...
func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	if !h.headersAreValid(w, r) {
		return
	}

	// recieve json body
	var reqBody UpdateExpenseRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		h.sendErrors(w, http.StatusBadRequest, []string{"unable to decode body"})
		return
	}

	// defer closing body
	defer func() {
		err = r.Body.Close()
		if err != nil {
			h.sendErrors(w, http.StatusInternalServerError, []string{})
		}
	}()

	// validatin of request body
	issues, isValid := reqBody.validate()
	if !isValid {
		h.sendErrors(w, http.StatusBadRequest, issues)
		return
	}

	// send to service layer
	err = h.Service.UpdateExpense(r.Context(), reqBody.ID, reqBody.OccuredAt.Time, reqBody.Description, reqBody.Amount)
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
		if errors.Is(err, expenses.ErrUnusedID) {
			h.sendErrors(w, http.StatusNotFound, []string{err.Error()})
		}

		// generic errors
		h.sendErrors(w, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// otherwise everything went perfect
	w.WriteHeader(http.StatusNoContent)
}

// DeleteExpense ...
func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idInt, err := validateID(idStr)
	if err != nil {
		h.sendErrors(w, http.StatusBadRequest, []string{err.Error()})
		return
	}

	// send to service
	err = h.Service.DeleteExpense(r.Context(), idInt)
	if err != nil {
		if errors.Is(err, expenses.ErrInvalidID) {
			h.sendErrors(w, http.StatusBadRequest, []string{err.Error()})
			return
		}
		if errors.Is(err, expenses.ErrUnusedID) {
			h.sendErrors(w, http.StatusNotFound, []string{err.Error()})
			return
		}

		// generic server error
		h.sendErrors(w, http.StatusInternalServerError, []string{})
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SummarizeExpenses ...
func (h *ExpenseHandler) SummarizeExpenses(w http.ResponseWriter, r *http.Request) {
	log.Println("summarize expenses is not implmeneted yet")
}
