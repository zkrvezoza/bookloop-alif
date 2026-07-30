package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/response"
	"github.com/bookloop-alif/internal/service"
)

type LoanManageHandler struct {
	loans *service.LoanService
	stats *service.StatsService
}

type extendRequest struct {
	Days int `json:"days" binding:"required"`
}

func NewLoanManageHandler(loans *service.LoanService, stats *service.StatsService) *LoanManageHandler {
	return &LoanManageHandler{loans: loans, stats: stats}
}

func (h *LoanManageHandler) List(c *gin.Context) {
	var userID int64
	if uidStr := c.Query("user"); uidStr != "" {
		id, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		userID = id
	}

	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	loans, total, err := h.loans.List(c.Request.Context(), repository.LoanFilter{
		UserID: userID,
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

func (h *LoanManageHandler) Stats(c *gin.Context) {
	stats, err := h.stats.Get(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.JSON(c, http.StatusOK, stats)
}

func (h *LoanManageHandler) Return(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	if err := h.loans.Return(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *LoanManageHandler) Extend(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	var req extendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.loans.ExtendDueDate(c.Request.Context(), id, req.Days); err != nil {
		response.Error(c, err)
		return
	}

	c.Status(http.StatusOK)
}
