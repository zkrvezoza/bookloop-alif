package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func performRequest(
	t *testing.T,
	method string,
	path string,
	body string,
	registerRoute func(*gin.Engine),
) *httptest.ResponseRecorder {
	t.Helper()

	router := gin.New()
	registerRoute(router)

	req := httptest.NewRequest(method, path, strings.NewReader(body))

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func requireStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()

	if rec.Code != expected {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			expected,
			rec.Code,
			rec.Body.String(),
		)
	}
}

// auth

func TestAuthHandler_Register_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/auth/register",
		`{}`,
		func(router *gin.Engine) {
			router.POST("/auth/register", handler.Register)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/auth/login",
		`{}`,
		func(router *gin.Engine) {
			router.POST("/auth/login", handler.Login)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestAuthHandler_Refresh_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/auth/refresh",
		`{}`,
		func(router *gin.Engine) {
			router.POST("/auth/refresh", handler.Refresh)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestAuthHandler_Logout_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/auth/logout",
		`{}`,
		func(router *gin.Engine) {
			router.POST("/auth/logout", handler.Logout)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

// bookmanage

func TestBookManageHandler_Create_InvalidBody(t *testing.T) {
	handler := NewBookManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/books",
		`{}`,
		func(router *gin.Engine) {
			router.POST("/books", handler.Create)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestBookManageHandler_Update_InvalidID(t *testing.T) {
	handler := NewBookManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPut,
		"/books/abc",
		`{}`,
		func(router *gin.Engine) {
			router.PUT("/books/:id", handler.Update)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestBookManageHandler_Update_InvalidBody(t *testing.T) {
	handler := NewBookManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPut,
		"/books/1",
		`{}`,
		func(router *gin.Engine) {
			router.PUT("/books/:id", handler.Update)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestBookManageHandler_Delete_InvalidID(t *testing.T) {
	handler := NewBookManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodDelete,
		"/books/abc",
		"",
		func(router *gin.Engine) {
			router.DELETE("/books/:id", handler.Delete)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestBookManageHandler_UploadCover_InvalidID(t *testing.T) {
	handler := NewBookManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/books/abc/cover",
		"",
		func(router *gin.Engine) {
			router.POST("/books/:id/cover", handler.UploadCover)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestBookManageHandler_UploadCover_MissingFile(t *testing.T) {
	handler := NewBookManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/books/1/cover",
		"",
		func(router *gin.Engine) {
			router.POST("/books/:id/cover", handler.UploadCover)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestBookManageHandler_GetCover_InvalidID(t *testing.T) {
	handler := NewBookManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodGet,
		"/books/abc/cover",
		"",
		func(router *gin.Engine) {
			router.GET("/books/:id/cover", handler.GetCover)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestBookManageHandler_MarkLost_InvalidID(t *testing.T) {
	handler := NewBookManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/loans/abc/lost",
		"",
		func(router *gin.Engine) {
			router.POST("/loans/:id/lost", handler.MarkLost)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

//bookreader

func TestBookReaderHandler_Get_InvalidID(t *testing.T) {
	handler := NewBookReaderHandler(nil)

	rec := performRequest(
		t,
		http.MethodGet,
		"/catalog/abc",
		"",
		func(router *gin.Engine) {
			router.GET("/catalog/:id", handler.Get)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

// loan manager

func TestLoanManageHandler_List_InvalidUserID(t *testing.T) {
	handler := NewLoanManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodGet,
		"/loans?user=abc",
		"",
		func(router *gin.Engine) {
			router.GET("/loans", handler.List)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestLoanManageHandler_Return_InvalidID(t *testing.T) {
	handler := NewLoanManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/loans/abc/return",
		"",
		func(router *gin.Engine) {
			router.POST("/loans/:id/return", handler.Return)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestLoanManageHandler_Extend_InvalidID(t *testing.T) {
	handler := NewLoanManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/loans/abc/extend",
		`{"days":7}`,
		func(router *gin.Engine) {
			router.POST("/loans/:id/extend", handler.Extend)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestLoanManageHandler_Extend_InvalidBody(t *testing.T) {
	handler := NewLoanManageHandler(nil, nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/loans/1/extend",
		`{}`,
		func(router *gin.Engine) {
			router.POST("/loans/:id/extend", handler.Extend)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

// loanreader

func TestLoanReaderHandler_Borrow_InvalidBookID(t *testing.T) {
	handler := NewLoanReaderHandler(nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/books/abc/borrow",
		"",
		func(router *gin.Engine) {
			router.POST("/books/:id/borrow", handler.Borrow)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestLoanReaderHandler_Return_InvalidLoanID(t *testing.T) {
	handler := NewLoanReaderHandler(nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/my-loans/abc/return",
		"",
		func(router *gin.Engine) {
			router.POST("/my-loans/:id/return", handler.Return)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

//reservation

func TestReservationHandler_Reserve_InvalidBookID(t *testing.T) {
	handler := NewReservationHandler(nil)

	rec := performRequest(
		t,
		http.MethodPost,
		"/books/abc/reserve",
		"",
		func(router *gin.Engine) {
			router.POST("/books/:id/reserve", handler.Reserve)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}

func TestReservationHandler_Cancel_InvalidID(t *testing.T) {
	handler := NewReservationHandler(nil)

	rec := performRequest(
		t,
		http.MethodDelete,
		"/reservations/abc",
		"",
		func(router *gin.Engine) {
			router.DELETE("/reservations/:id", handler.Cancel)
		},
	)

	requireStatus(t, rec, http.StatusBadRequest)
}
