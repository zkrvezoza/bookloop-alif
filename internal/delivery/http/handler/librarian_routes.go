package handler

import "github.com/gin-gonic/gin"

func RegisterLibrarianRoutes(rg *gin.RouterGroup, book *BookManageHandler, loan *LoanManageHandler) {
	rg.POST("/books", book.Create)
	rg.PUT("/books/:id", book.Update)
	rg.DELETE("/books/:id", book.Delete)
	rg.POST("/books/:id/cover", book.UploadCover)
	rg.GET("/books/:id/cover", book.GetCover)
	rg.PUT("/books/:id/lost", book.MarkLost)

	rg.GET("/loans", loan.List)
	rg.PUT("/loans/:id/return", loan.Return)
	rg.PUT("/loans/:id/extend", loan.Extend)
	rg.GET("/stats", loan.Stats)
}
