package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/jackc/pgx/v5"
)

type LoanRepo struct {
	db Querier
}

func NewLoanRepo(db Querier) *LoanRepo {
	return &LoanRepo{db: db}
}

func (r *LoanRepo) Create(ctx context.Context, l *model.Loan) (*model.Loan, error) {
	const q = `
		INSERT INTO loans (book_id, user_id, due_at, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, borrowed_at`

	err := r.db.QueryRow(ctx, q, l.BookID, l.UserID, l.DueAt, l.Status).
		Scan(&l.ID, &l.BorrowedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return l, nil
}

func (r *LoanRepo) GetByID(ctx context.Context, id int64) (*model.Loan, error) {
	const q = `
		SELECT id, book_id, user_id, borrowed_at, due_at, returned_at, status
		FROM loans WHERE id = $1`

	l := &model.Loan{}
	err := r.db.QueryRow(ctx, q, id).
		Scan(&l.ID, &l.BookID, &l.UserID, &l.BorrowedAt, &l.DueAt, &l.ReturnedAt, &l.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *LoanRepo) List(ctx context.Context, f repository.LoanFilter) ([]model.Loan, int, error) {
	where := []string{"1=1"}
	args := []any{}
	argN := 1

	if f.UserID != 0 {
		where = append(where, fmt.Sprintf("user_id = $%d", argN))
		args = append(args, f.UserID)
		argN++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, f.Status)
		argN++
	}
	whereClause := ""
	for i, cond := range where {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += cond
	}

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM loans WHERE "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, page := f.Limit, f.Page
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	listQ := fmt.Sprintf(`
		SELECT id, book_id, user_id, borrowed_at, due_at, returned_at, status
		FROM loans WHERE %s
		ORDER BY id
		LIMIT $%d OFFSET $%d`, whereClause, argN, argN+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var loans []model.Loan
	for rows.Next() {
		var l model.Loan
		if err := rows.Scan(&l.ID, &l.BookID, &l.UserID, &l.BorrowedAt, &l.DueAt, &l.ReturnedAt, &l.Status); err != nil {
			return nil, 0, err
		}
		loans = append(loans, l)
	}
	return loans, total, rows.Err()
}

func (r *LoanRepo) MarkReturned(ctx context.Context, id int64) error {
	const q = `UPDATE loans SET status = 'returned', returned_at = NOW() WHERE id = $1 AND status = 'active'`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *LoanRepo) MarkOverdue(ctx context.Context, id int64) error {
	const q = `UPDATE loans SET status = 'overdue' WHERE id = $1 AND status = 'active'`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *LoanRepo) ExtendDueDate(ctx context.Context, id int64, newDueAt string) error {
	const q = `UPDATE loans SET due_at = $1 WHERE id = $2 AND status = 'active'`
	tag, err := r.db.Exec(ctx, q, newDueAt, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *LoanRepo) CountActiveByBook(ctx context.Context, bookID int64) (int, error) {
	const q = `SELECT COUNT(*) FROM loans WHERE book_id = $1 AND status = 'active'`
	var count int
	err := r.db.QueryRow(ctx, q, bookID).Scan(&count)
	return count, err
}
