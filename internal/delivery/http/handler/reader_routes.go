package handler

import "github.com/gin-gonic/gin"

func RegisterReaderRoutes(rg *gin.RouterGroup, book *BookReaderHandler, loan *LoanReaderHandler, res *ReservationHandler) {
	rg.GET("/books", book.List)
	rg.GET("/books/:id", book.Get)

	rg.POST("/books/:id/loan", loan.Borrow)
	rg.GET("/loans", loan.MyLoans)
	rg.PUT("/loans/:id/return", loan.Return)

	rg.POST("/books/:id/reserve", res.Reserve)
	rg.GET("/reservations", res.MyReservations)
	rg.DELETE("/reservations/:id", res.Cancel)
}
