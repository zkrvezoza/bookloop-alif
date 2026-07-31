package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/service"
)

// ---- fake repos (то же самое, что в internal/service тестах, но локально для этого пакета) ----

type fakeUserRepo struct {
	byLogin map[string]*model.User
	nextID  int64
}

func newFakeUserRepo() *fakeUserRepo { return &fakeUserRepo{byLogin: make(map[string]*model.User)} }

func (f *fakeUserRepo) Create(_ context.Context, u *model.User) (*model.User, error) {
	if _, exists := f.byLogin[u.Login]; exists {
		return nil, model.ErrInvalid
	}
	f.nextID++
	u.ID = f.nextID
	f.byLogin[u.Login] = u
	return u, nil
}
func (f *fakeUserRepo) GetByLogin(_ context.Context, login string) (*model.User, error) {
	u, ok := f.byLogin[login]
	if !ok {
		return nil, model.ErrNotFound
	}
	return u, nil
}
func (f *fakeUserRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	for _, u := range f.byLogin {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, model.ErrNotFound
}

type fakeRefreshTokenRepo struct {
	tokens map[string]*model.RefreshToken
	nextID int64
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{tokens: make(map[string]*model.RefreshToken)}
}
func (f *fakeRefreshTokenRepo) Create(_ context.Context, rt *model.RefreshToken) (*model.RefreshToken, error) {
	f.nextID++
	rt.ID = f.nextID
	f.tokens[rt.TokenHash] = rt
	return rt, nil
}
func (f *fakeRefreshTokenRepo) GetByHash(_ context.Context, hash string) (*model.RefreshToken, error) {
	rt, ok := f.tokens[hash]
	if !ok {
		return nil, model.ErrNotFound
	}
	return rt, nil
}
func (f *fakeRefreshTokenRepo) Revoke(_ context.Context, id int64) error {
	for _, rt := range f.tokens {
		if rt.ID == id {
			now := rt.ExpiresAt
			rt.RevokedAt = &now
			return nil
		}
	}
	return model.ErrNotFound
}

type fakeBookRepo struct {
	books  map[int64]*model.Book
	nextID int64
}

func newFakeBookRepo() *fakeBookRepo { return &fakeBookRepo{books: make(map[int64]*model.Book)} }

func (f *fakeBookRepo) Create(_ context.Context, b *model.Book) (*model.Book, error) {
	f.nextID++
	b.ID = f.nextID
	f.books[b.ID] = b
	return b, nil
}
func (f *fakeBookRepo) GetByID(_ context.Context, id int64) (*model.Book, error) {
	b, ok := f.books[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return b, nil
}
func (f *fakeBookRepo) Update(_ context.Context, b *model.Book) error {
	if _, ok := f.books[b.ID]; !ok {
		return model.ErrNotFound
	}
	f.books[b.ID] = b
	return nil
}
func (f *fakeBookRepo) Delete(_ context.Context, id int64) error {
	if _, ok := f.books[id]; !ok {
		return model.ErrNotFound
	}
	delete(f.books, id)
	return nil
}
func (f *fakeBookRepo) List(_ context.Context, filter repository.BookFilter) ([]model.Book, int, error) {
	var out []model.Book
	for _, b := range f.books {
		if filter.Genre != "" && b.Genre != filter.Genre {
			continue
		}
		out = append(out, *b)
	}
	return out, len(out), nil
}
func (f *fakeBookRepo) DecrCopies(_ context.Context, id int64) error {
	b, ok := f.books[id]
	if !ok {
		return model.ErrNotFound
	}
	if b.Copies <= 0 {
		return model.ErrNoSeats
	}
	b.Copies--
	return nil
}
func (f *fakeBookRepo) IncrCopies(_ context.Context, id int64) error {
	b, ok := f.books[id]
	if !ok {
		return model.ErrNotFound
	}
	b.Copies++
	return nil
}
func (f *fakeBookRepo) MarkLost(_ context.Context, id int64) error {
	b, ok := f.books[id]
	if !ok {
		return model.ErrNotFound
	}
	b.Status = model.BookLost
	return nil
}

type fakeLoanRepo struct {
	loans  map[int64]*model.Loan
	nextID int64
}

func newFakeLoanRepo() *fakeLoanRepo { return &fakeLoanRepo{loans: make(map[int64]*model.Loan)} }

func (f *fakeLoanRepo) Create(_ context.Context, l *model.Loan) (*model.Loan, error) {
	f.nextID++
	l.ID = f.nextID
	f.loans[l.ID] = l
	return l, nil
}
func (f *fakeLoanRepo) GetByID(_ context.Context, id int64) (*model.Loan, error) {
	l, ok := f.loans[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return l, nil
}
func (f *fakeLoanRepo) List(_ context.Context, filter repository.LoanFilter) ([]model.Loan, int, error) {
	var out []model.Loan
	for _, l := range f.loans {
		if filter.UserID != 0 && l.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && string(l.Status) != filter.Status {
			continue
		}
		out = append(out, *l)
	}
	return out, len(out), nil
}
func (f *fakeLoanRepo) MarkReturned(_ context.Context, id int64) error {
	l, ok := f.loans[id]
	if !ok || l.Status != model.LoanActive {
		return model.ErrNotFound
	}
	l.Status = model.LoanReturned
	return nil
}
func (f *fakeLoanRepo) MarkOverdue(_ context.Context, id int64) error {
	l, ok := f.loans[id]
	if !ok || l.Status != model.LoanActive {
		return model.ErrNotFound
	}
	l.Status = model.LoanOverdue
	return nil
}
func (f *fakeLoanRepo) ExtendDueDate(_ context.Context, id int64, _ string) error {
	l, ok := f.loans[id]
	if !ok || l.Status != model.LoanActive {
		return model.ErrNotFound
	}
	return nil
}
func (f *fakeLoanRepo) CountActiveByBook(_ context.Context, bookID int64) (int, error) {
	count := 0
	for _, l := range f.loans {
		if l.BookID == bookID && l.Status == model.LoanActive {
			count++
		}
	}
	return count, nil
}

type fakeReservationRepo struct {
	res    map[int64]*model.Reservation
	nextID int64
}

func newFakeReservationRepo() *fakeReservationRepo {
	return &fakeReservationRepo{res: make(map[int64]*model.Reservation)}
}
func (f *fakeReservationRepo) Create(_ context.Context, r *model.Reservation) (*model.Reservation, error) {
	f.nextID++
	r.ID = f.nextID
	f.res[r.ID] = r
	return r, nil
}
func (f *fakeReservationRepo) NextWaiting(_ context.Context, bookID int64) (*model.Reservation, error) {
	var found *model.Reservation
	for _, r := range f.res {
		if r.BookID == bookID && r.Status == model.ReservationWaiting {
			if found == nil || r.ID < found.ID {
				found = r
			}
		}
	}
	if found == nil {
		return nil, model.ErrNotFound
	}
	return found, nil
}
func (f *fakeReservationRepo) MarkFulfilled(_ context.Context, id int64) error {
	r, ok := f.res[id]
	if !ok || r.Status != model.ReservationWaiting {
		return model.ErrNotFound
	}
	r.Status = model.ReservationFulfilled
	return nil
}
func (f *fakeReservationRepo) Cancel(_ context.Context, id int64) error {
	r, ok := f.res[id]
	if !ok {
		return model.ErrNotFound
	}
	r.Status = model.ReservationCancelled
	return nil
}
func (f *fakeReservationRepo) ListByUser(_ context.Context, userID int64) ([]model.Reservation, error) {
	var out []model.Reservation
	for _, r := range f.res {
		if r.UserID == userID {
			out = append(out, *r)
		}
	}
	return out, nil
}

type fakeStatsRepo struct{ row *repository.LibraryStatsRow }

func (f *fakeStatsRepo) LibraryStats(_ context.Context) (*repository.LibraryStatsRow, error) {
	return f.row, nil
}

// fakeUOW — не открывает реальную транзакцию, просто вызывает fn с теми же fake-репозиториями.
type fakeUOW struct {
	books *fakeBookRepo
	loans *fakeLoanRepo
}

func (u *fakeUOW) WithinTx(ctx context.Context, fn func(books repository.BookRepo, loans repository.LoanRepo) error) error {
	return fn(u.books, u.loans)
}

// withUser — тестовый middleware, кладёт userID в контекст так же, как это делает middleware.Auth.
func withUser(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// ---- AuthHandler happy-path ----

func TestAuthHandler_RegisterAndLogin_Success(t *testing.T) {
	users := newFakeUserRepo()
	tokens := newFakeRefreshTokenRepo()
	authSvc := service.NewAuthService(users, tokens, "test-secret")
	h := NewAuthHandler(authSvc)

	rec := performRequest(t, http.MethodPost, "/auth/register",
		`{"login":"aziza","password":"password123","role":"user"}`,
		func(r *gin.Engine) { r.POST("/auth/register", h.Register) })
	requireStatus(t, rec, http.StatusCreated)

	rec = performRequest(t, http.MethodPost, "/auth/login",
		`{"login":"aziza","password":"password123"}`,
		func(r *gin.Engine) { r.POST("/auth/login", h.Login) })
	requireStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), "access_token") {
		t.Fatalf("expected access_token in response, got: %s", rec.Body.String())
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	users := newFakeUserRepo()
	tokens := newFakeRefreshTokenRepo()
	authSvc := service.NewAuthService(users, tokens, "test-secret")
	h := NewAuthHandler(authSvc)

	performRequest(t, http.MethodPost, "/auth/register",
		`{"login":"aziza","password":"password123","role":"user"}`,
		func(r *gin.Engine) { r.POST("/auth/register", h.Register) })

	rec := performRequest(t, http.MethodPost, "/auth/login",
		`{"login":"aziza","password":"wrong"}`,
		func(r *gin.Engine) { r.POST("/auth/login", h.Login) })
	requireStatus(t, rec, http.StatusUnauthorized)
}

// ---- BookManageHandler happy-path ----

func newTestBookManageHandler() (*BookManageHandler, *fakeBookRepo, *fakeLoanRepo) {
	books := newFakeBookRepo()
	loans := newFakeLoanRepo()
	res := newFakeReservationRepo()
	uow := &fakeUOW{books: books, loans: loans}

	bookSvc := service.NewBookService(books, loans)
	loanSvc := service.NewLoanService(loans, books, res, uow)
	return NewBookManageHandler(bookSvc, loanSvc), books, loans
}

func TestBookManageHandler_Create_Success(t *testing.T) {
	h, _, _ := newTestBookManageHandler()

	rec := performRequest(t, http.MethodPost, "/books",
		`{"title":"Dune","author":"Herbert","genre":"scifi","copies":3}`,
		func(r *gin.Engine) { r.POST("/books", h.Create) })

	requireStatus(t, rec, http.StatusCreated)
}

func TestBookManageHandler_List_Success(t *testing.T) {
	h, books, _ := newTestBookManageHandler()
	_, err := books.Create(context.Background(), &model.Book{
		Title:  "Dune",
		Genre:  "scifi",
		Copies: 1,
		Status: model.BookAvailable,
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := performRequest(t, http.MethodGet, "/books", "",
		func(r *gin.Engine) { r.GET("/books", h.List) })

	requireStatus(t, rec, http.StatusOK)
}

func TestBookManageHandler_Update_Success(t *testing.T) {
	h, books, _ := newTestBookManageHandler()
	b, _ := books.Create(context.Background(), &model.Book{Title: "Dune", Genre: "scifi", Copies: 5, Status: model.BookAvailable})

	rec := performRequest(t, http.MethodPut, "/books/1",
		`{"title":"Dune 2","author":"H","genre":"scifi","copies":3}`,
		func(r *gin.Engine) { r.PUT("/books/:id", h.Update) })

	requireStatus(t, rec, http.StatusOK)
	if books.books[b.ID].Title != "Dune 2" {
		t.Fatalf("update didn't apply")
	}
}

func TestBookManageHandler_Delete_Success(t *testing.T) {
	h, books, _ := newTestBookManageHandler()
	_, err := books.Create(context.Background(), &model.Book{
		Title:  "Dune",
		Genre:  "scifi",
		Copies: 1,
		Status: model.BookAvailable,
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := performRequest(t, http.MethodDelete, "/books/1", "",
		func(r *gin.Engine) { r.DELETE("/books/:id", h.Delete) })

	requireStatus(t, rec, http.StatusNoContent)
}

func TestBookManageHandler_GetCover_NotFound(t *testing.T) {
	h, books, _ := newTestBookManageHandler()
	_, err := books.Create(context.Background(), &model.Book{
		Title:  "Dune",
		Genre:  "scifi",
		Copies: 1,
		Status: model.BookAvailable,
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := performRequest(t, http.MethodGet, "/books/1/cover", "",
		func(r *gin.Engine) { r.GET("/books/:id/cover", h.GetCover) })

	requireStatus(t, rec, http.StatusNotFound)
}

func TestBookManageHandler_MarkLost_Success(t *testing.T) {
	h, books, loans := newTestBookManageHandler()
	b, _ := books.Create(context.Background(), &model.Book{Title: "Dune", Genre: "scifi", Copies: 1, Status: model.BookAvailable})
	if _, err := loans.Create(context.Background(), &model.Loan{
		BookID: b.ID,
		UserID: 1,
		Status: model.LoanActive,
	}); err != nil {
		t.Fatal(err)
	}
	rec := performRequest(t, http.MethodPost, "/loans/1/lost", "",
		func(r *gin.Engine) { r.POST("/loans/:id/lost", h.MarkLost) })

	requireStatus(t, rec, http.StatusOK)
	if books.books[b.ID].Status != model.BookLost {
		t.Fatalf("book should be marked lost")
	}
}

// ---- BookReaderHandler happy-path ----

func TestBookReaderHandler_List_Success(t *testing.T) {
	books := newFakeBookRepo()

	if _, err := books.Create(context.Background(), &model.Book{
		Title:  "T",
		Copies: 0,
		Status: model.BookAvailable,
	}); err != nil {
		t.Fatal(err)
	}
	loans := newFakeLoanRepo()
	h := NewBookReaderHandler(service.NewBookService(books, loans))

	rec := performRequest(t, http.MethodGet, "/books", "",
		func(r *gin.Engine) { r.GET("/books", h.List) })

	requireStatus(t, rec, http.StatusOK)
}

func TestBookReaderHandler_Get_NotFound(t *testing.T) {
	h := NewBookReaderHandler(service.NewBookService(newFakeBookRepo(), newFakeLoanRepo()))

	rec := performRequest(t, http.MethodGet, "/books/999", "",
		func(r *gin.Engine) { r.GET("/books/:id", h.Get) })

	requireStatus(t, rec, http.StatusNotFound)
}

// ---- LoanManageHandler happy-path ----

func newTestLoanManageHandler() (*LoanManageHandler, *fakeLoanRepo, *fakeBookRepo) {
	books := newFakeBookRepo()
	loans := newFakeLoanRepo()
	res := newFakeReservationRepo()
	uow := &fakeUOW{books: books, loans: loans}
	stats := &fakeStatsRepo{row: &repository.LibraryStatsRow{BooksTotal: 1}}

	loanSvc := service.NewLoanService(loans, books, res, uow)
	statsSvc := service.NewStatsService(stats)
	return NewLoanManageHandler(loanSvc, statsSvc), loans, books
}

func TestLoanManageHandler_List_Success(t *testing.T) {
	h, loans, _ := newTestLoanManageHandler()
	if _, err := loans.Create(context.Background(), &model.Loan{
		BookID: 1,
		UserID: 1,
		Status: model.LoanActive,
	}); err != nil {
		t.Fatal(err)
	}
	rec := performRequest(t, http.MethodGet, "/loans", "",
		func(r *gin.Engine) { r.GET("/loans", h.List) })

	requireStatus(t, rec, http.StatusOK)
}

func TestLoanManageHandler_Stats_Success(t *testing.T) {
	h, _, _ := newTestLoanManageHandler()

	rec := performRequest(t, http.MethodGet, "/stats", "",
		func(r *gin.Engine) { r.GET("/stats", h.Stats) })

	requireStatus(t, rec, http.StatusOK)
}

func TestLoanManageHandler_Return_Success(t *testing.T) {
	h, loans, books := newTestLoanManageHandler()
	b, _ := books.Create(context.Background(), &model.Book{Title: "T", Copies: 0, Status: model.BookAvailable})
	l, _ := loans.Create(context.Background(), &model.Loan{BookID: b.ID, UserID: 1, Status: model.LoanActive})

	rec := performRequest(t, http.MethodPut, "/loans/1/return", "",
		func(r *gin.Engine) { r.PUT("/loans/:id/return", h.Return) })

	requireStatus(t, rec, http.StatusOK)
	if loans.loans[l.ID].Status != model.LoanReturned {
		t.Fatalf("loan should be returned")
	}
}

func TestLoanManageHandler_Extend_Success(t *testing.T) {
	h, loans, _ := newTestLoanManageHandler()
	if _, err := loans.Create(context.Background(), &model.Loan{
		BookID: 1,
		UserID: 1,
		Status: model.LoanActive,
	}); err != nil {
		t.Fatal(err)
	}
	rec := performRequest(t, http.MethodPut, "/loans/1/extend", `{"days":7}`,
		func(r *gin.Engine) { r.PUT("/loans/:id/extend", h.Extend) })

	requireStatus(t, rec, http.StatusOK)
}

// ---- LoanReaderHandler happy-path ----

func newTestLoanReaderHandler() (*LoanReaderHandler, *fakeBookRepo, *fakeLoanRepo) {
	books := newFakeBookRepo()
	loans := newFakeLoanRepo()
	res := newFakeReservationRepo()
	uow := &fakeUOW{books: books, loans: loans}
	loanSvc := service.NewLoanService(loans, books, res, uow)
	return NewLoanReaderHandler(loanSvc), books, loans
}

func TestLoanReaderHandler_Borrow_Success(t *testing.T) {
	h, books, _ := newTestLoanReaderHandler()
	if _, err := books.Create(context.Background(), &model.Book{
		Title:  "T",
		Copies: 1,
		Status: model.BookAvailable,
	}); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(withUser(1))
	router.POST("/books/:id/borrow", h.Borrow)

	req := httptest.NewRequest(http.MethodPost, "/books/1/borrow", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	requireStatus(t, rec, http.StatusCreated)
}

func TestLoanReaderHandler_Borrow_NoSeats(t *testing.T) {
	h, books, _ := newTestLoanReaderHandler()
	if _, err := books.Create(context.Background(), &model.Book{
		Title:  "T",
		Copies: 0,
		Status: model.BookAvailable,
	}); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(withUser(1))
	router.POST("/books/:id/borrow", h.Borrow)

	req := httptest.NewRequest(http.MethodPost, "/books/1/borrow", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	requireStatus(t, rec, http.StatusConflict)
}

func TestLoanReaderHandler_Return_ForbiddenForOtherUser(t *testing.T) {
	h, books, loans := newTestLoanReaderHandler()
	b, _ := books.Create(context.Background(), &model.Book{Title: "T", Copies: 0, Status: model.BookAvailable})
	if _, err := loans.Create(context.Background(), &model.Loan{
		BookID: b.ID,
		UserID: 1,
		Status: model.LoanActive,
	}); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(withUser(2)) // другой пользователь
	router.POST("/loans/:id/return", h.Return)

	req := httptest.NewRequest(http.MethodPost, "/loans/1/return", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	requireStatus(t, rec, http.StatusForbidden)
}

// ---- ReservationHandler happy-path ----

func TestReservationHandler_Reserve_Success(t *testing.T) {
	books := newFakeBookRepo()
	if _, err := books.Create(context.Background(), &model.Book{
		Title:  "T",
		Copies: 0,
		Status: model.BookAvailable,
	}); err != nil {
		t.Fatal(err)
	}
	res := newFakeReservationRepo()
	h := NewReservationHandler(service.NewReservationService(res, books))

	router := gin.New()
	router.Use(withUser(1))
	router.POST("/books/:id/reserve", h.Reserve)

	req := httptest.NewRequest(http.MethodPost, "/books/1/reserve", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	requireStatus(t, rec, http.StatusCreated)
}

func TestReservationHandler_MyReservations_Success(t *testing.T) {
	books := newFakeBookRepo()
	res := newFakeReservationRepo()
	if _, err := res.Create(context.Background(), &model.Reservation{
		BookID: 1,
		UserID: 1,
		Status: model.ReservationWaiting,
	}); err != nil {
		t.Fatal(err)
	}
	h := NewReservationHandler(service.NewReservationService(res, books))

	router := gin.New()
	router.Use(withUser(1))
	router.GET("/reservations", h.MyReservations)

	req := httptest.NewRequest(http.MethodGet, "/reservations", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	requireStatus(t, rec, http.StatusOK)
}

func TestReservationHandler_Cancel_Success(t *testing.T) {
	books := newFakeBookRepo()
	res := newFakeReservationRepo()
	if _, err := res.Create(context.Background(), &model.Reservation{
		BookID: 1,
		UserID: 1,
		Status: model.ReservationWaiting,
	}); err != nil {
		t.Fatal(err)
	}
	h := NewReservationHandler(service.NewReservationService(res, books))

	router := gin.New()
	router.Use(withUser(1))
	router.DELETE("/reservations/:id", h.Cancel)

	req := httptest.NewRequest(http.MethodDelete, "/reservations/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	requireStatus(t, rec, http.StatusNoContent)
}
