package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/delivery/http/middleware"
	"github.com/bookloop-alif/internal/response"
	"github.com/bookloop-alif/internal/service"
)

type ReservationHandler struct {
	res *service.ReservationService
}

func NewReservationHandler(res *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{res: res}
}

func (h *ReservationHandler) Reserve(c *gin.Context) {
	bookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	userID := middleware.UserID(c)

	r, err := h.res.Reserve(c.Request.Context(), userID, bookID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, r)
}

func (h *ReservationHandler) MyReservations(c *gin.Context) {
	userID := middleware.UserID(c)

	list, err := h.res.MyReservations(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, list)
}

func (h *ReservationHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation id"})
		return
	}

	userID := middleware.UserID(c)

	if err := h.res.Cancel(c.Request.Context(), userID, id); err != nil {
		response.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
