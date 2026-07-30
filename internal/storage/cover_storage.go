package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/bookloop-alif/internal/domain/model"
)

const (
	maxCoverSizeBytes = 5 << 20
	coversDir         = "uploads/covers"
)

var allowedExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

func SaveCover(bookID int64, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > maxCoverSizeBytes {
		return "", model.ErrInvalid
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExt[ext] {
		return "", model.ErrInvalid
	}

	if err := os.MkdirAll(coversDir, 0o755); err != nil {
		return "", err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()

	relPath := filepath.Join(coversDir, fmt.Sprintf("book_%d%s", bookID, ext))
	dst, err := os.Create(relPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return relPath, nil
}
