package handler

import "github.com/gin-gonic/gin"

func RegisterLibrarianRoutes(rg *gin.RouterGroup, h *BookManageHandler) {
	rg.POST("/books", h.Create)
	rg.PUT("/books/:id", h.Update)
	rg.DELETE("/books/:id", h.Delete)
	rg.POST("/books/:id/cover", h.UploadCover)
	rg.GET("/books/:id/cover", h.GetCover)
}
