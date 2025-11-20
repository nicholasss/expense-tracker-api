package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/handler"
)

func SetupGinRoutes(service expenses.Service) *gin.Engine {
	h := handler.NewGinHandler(service)

	r := gin.Default()

	r.GET("/expenses", h.GetAllExpenses)
	// get expenses by id
	// post new expenses
	// put expenses
	// delete expenses

	return r
}
