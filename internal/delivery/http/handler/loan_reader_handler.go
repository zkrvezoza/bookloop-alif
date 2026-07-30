package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/delivery/http/middleware"
	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/response"
	"github.com/bookloop-alif/internal/service"
)

type LoanReaderHandler struct {
	loans *service.ReaderLoanService
}

func NewLoanReaderHandler(loans *service.ReaderLoanService) *LoanReaderHandler {
	return &LoanReaderHandler{loans: loans}
}

func (h *LoanReaderHandler) Borrow(c *gin.Context) {
	bookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	userID := middleware.UserID(c)

	loan, err := h.loans.Borrow(c.Request.Context(), userID, bookID)
	if err != nil {
		if errors.Is(err, model.ErrNoSeats) {
			c.JSON(http.StatusConflict, gin.H{"error": "no available copies, use reservation queue"})
			return
		}
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, loan)
}

func (h *LoanReaderHandler) MyLoans(c *gin.Context) {
	userID := middleware.UserID(c)

	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	loans, total, err := h.loans.MyLoans(c.Request.Context(), userID, repository.LoanFilter{
		Status: c.Query("status"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"items": loans, "total": total, "page": page})
}

func (h *LoanReaderHandler) Return(c *gin.Context) {
	loanID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	userID := middleware.UserID(c)

	if err := h.loans.Return(c.Request.Context(), userID, loanID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusOK)
}
