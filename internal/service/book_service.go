package service

import (
	"context"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
)

type BookService struct {
	books repository.BookRepo
	loans repository.LoanRepo
}

func NewBookService(books repository.BookRepo, loans repository.LoanRepo) *BookService {
	return &BookService{books: books, loans: loans}
}

func (s *BookService) Create(ctx context.Context, title, author, genre string, copies int) (*model.Book, error) {
	if title == "" || author == "" || genre == "" || copies <= 0 {
		return nil, model.ErrInvalid
	}

	return s.books.Create(ctx, &model.Book{
		Title:  title,
		Author: author,
		Genre:  genre,
		Status: model.BookAvailable,
		Copies: copies,
	})
}

func (s *BookService) Get(ctx context.Context, id int64) (*model.Book, error) {
	return s.books.GetByID(ctx, id)
}

type UpdateBookInput struct {
	Title  string
	Author string
	Genre  string
	Copies int
}

func (s *BookService) Update(ctx context.Context, id int64, in UpdateBookInput) error {
	b, err := s.books.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if in.Title == "" || in.Author == "" || in.Genre == "" || in.Copies <= 0 {
		return model.ErrInvalid
	}

	activeLoans, err := s.loans.CountActiveByBook(ctx, id)
	if err != nil {
		return err
	}
	if in.Copies < activeLoans {
		return model.ErrInvalid
	}

	b.Title = in.Title
	b.Author = in.Author
	b.Genre = in.Genre
	b.Copies = in.Copies

	return s.books.Update(ctx, b)
}

func (s *BookService) Delete(ctx context.Context, id int64) error {
	return s.books.Delete(ctx, id)
}

func (s *BookService) SetCoverPath(ctx context.Context, id int64, path string) error {
	b, err := s.books.GetByID(ctx, id)
	if err != nil {
		return err
	}
	b.CoverPath = path
	return s.books.Update(ctx, b)
}

func (s *BookService) MarkLost(ctx context.Context, id int64) error {
	b, err := s.books.GetByID(ctx, id)
	if err != nil {
		return err
	}
	b.Status = model.BookLost
	return s.books.Update(ctx, b)
}
