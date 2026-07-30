package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/response"
	"github.com/bookloop-alif/internal/service"
	"github.com/bookloop-alif/internal/storage"
)

type BookManageHandler struct {
	books *service.BookService
}

func NewBookManageHandler(books *service.BookService) *BookManageHandler {
	return &BookManageHandler{books: books}
}

type createBookRequest struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
	Genre  string `json:"genre" binding:"required"`
	Copies int    `json:"copies" binding:"required"`
}

func (h *BookManageHandler) Create(c *gin.Context) {
	var req createBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	b, err := h.books.Create(c.Request.Context(), req.Title, req.Author, req.Genre, req.Copies)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, b)
}

type updateBookRequest struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
	Genre  string `json:"genre" binding:"required"`
	Copies int    `json:"copies" binding:"required"`
}

func (h *BookManageHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err = h.books.Update(c.Request.Context(), id, service.UpdateBookInput{
		Title:  req.Title,
		Author: req.Author,
		Genre:  req.Genre,
		Copies: req.Copies,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *BookManageHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.books.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BookManageHandler) UploadCover(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	fileHeader, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cover file required"})
		return
	}

	path, err := storage.SaveCover(id, fileHeader)
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.books.SetCoverPath(c.Request.Context(), id, path); err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"cover_path": path})
}

func (h *BookManageHandler) GetCover(c *gin.Context) {
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
	if b.CoverPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no cover uploaded"})
		return
	}

	c.File(b.CoverPath)
}
