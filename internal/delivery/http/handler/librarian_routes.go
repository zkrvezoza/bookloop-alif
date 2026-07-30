package handler

import "github.com/gin-gonic/gin"

func RegisterLibrarianRoutes(rg *gin.RouterGroup, book *BookManageHandler, loan *LoanManageHandler) {
	rg.POST("/books", book.Create)
	rg.PUT("/books/:id", book.Update)
	rg.DELETE("/books/:id", book.Delete)
	rg.POST("/books/:id/cover", book.UploadCover)
	rg.GET("/books/:id/cover", book.GetCover)

	rg.GET("/loans", loan.List)
	rg.GET("/stats", loan.Stats)
}
