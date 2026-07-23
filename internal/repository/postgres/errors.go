package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/bookloop-alif/internal/domain/model"
)

func mapPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return model.ErrInvalid
		case "23503":
			return model.ErrInvalid
		}
	}
	return err
}
