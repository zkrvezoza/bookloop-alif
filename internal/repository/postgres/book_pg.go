package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/jackc/pgx/v5"
)

type BookRepo struct {
	db Querier
}

func NewBookRepo(db Querier) *BookRepo {
	return &BookRepo{db: db}
}

func (r *BookRepo) Create(ctx context.Context, b *model.Book) (*model.Book, error) {
	const q = `
		INSERT INTO books (title, author, genre, cover_path, status, copies)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q, b.Title, b.Author, b.Genre, b.CoverPath, b.Status, b.Copies).
		Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return b, nil
}

func (r *BookRepo) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	const q = `
		SELECT id, title, author, genre, cover_path, status, copies, created_at
		FROM books WHERE id = $1`

	b := &model.Book{}
	err := r.db.QueryRow(ctx, q, id).
		Scan(&b.ID, &b.Title, &b.Author, &b.Genre, &b.CoverPath, &b.Status, &b.Copies, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BookRepo) Update(ctx context.Context, b *model.Book) error {
	const q = `
		UPDATE books
		SET title = $1, author = $2, genre = $3, cover_path = $4, status = $5, copies = $6
		WHERE id = $7`

	tag, err := r.db.Exec(ctx, q, b.Title, b.Author, b.Genre, b.CoverPath, b.Status, b.Copies, b.ID)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *BookRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *BookRepo) List(ctx context.Context, f repository.BookFilter) ([]model.Book, int, error) {
	where := []string{"1=1"}
	args := []any{}
	argN := 1

	if f.Query != "" {
		where = append(where, fmt.Sprintf(
			"(title ILIKE $%d OR author ILIKE $%d OR genre ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+f.Query+"%")
		argN++
	}
	if f.Genre != "" {
		where = append(where, fmt.Sprintf("genre = $%d", argN))
		args = append(args, f.Genre)
		argN++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, f.Status)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQ := "SELECT COUNT(*) FROM books WHERE " + whereClause
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
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
		SELECT id, title, author, genre, cover_path, status, copies, created_at
		FROM books WHERE %s
		ORDER BY id
		LIMIT $%d OFFSET $%d`, whereClause, argN, argN+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var books []model.Book
	for rows.Next() {
		var b model.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Genre, &b.CoverPath, &b.Status, &b.Copies, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		books = append(books, b)
	}
	return books, total, rows.Err()
}

func (r *BookRepo) DecrCopies(ctx context.Context, id int64) error {
	const q = `UPDATE books SET copies = copies - 1 WHERE id = $1 AND copies > 0`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNoSeats
	}
	return nil
}

func (r *BookRepo) IncrCopies(ctx context.Context, id int64) error {
	const q = `UPDATE books SET copies = copies + 1 WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *BookRepo) MarkLost(ctx context.Context, id int64) error {
	const q = `UPDATE books SET status = 'lost' WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
