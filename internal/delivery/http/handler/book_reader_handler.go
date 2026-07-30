package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/response"
	"github.com/bookloop-alif/internal/service"
)

type BookReaderHandler struct {
	books *service.BookService
}

func NewBookReaderHandler(books *service.BookService) *BookReaderHandler {
	return &BookReaderHandler{books: books}
}

// List — A-01 (каталог+поиск) + A-05 (пагинация/фильтры).
func (h *BookReaderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	books, total, err := h.books.List(c.Request.Context(), repository.BookFilter{
		Query:  c.Query("q"),
		Genre:  c.Query("genre"),
		Status: c.Query("status"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"items": books, "total": total, "page": page})
}

func (h *BookReaderHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	b, err := h.books.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.JSON(c, http.StatusOK, b)
}
