package storage

import (
	"bytes"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
)

func newFileHeader(
	t *testing.T,
	filename string,
	content []byte,
) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("cover", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	reader := multipart.NewReader(
		bytes.NewReader(body.Bytes()),
		writer.Boundary(),
	)

	form, err := reader.ReadForm(int64(len(body.Bytes())))
	if err != nil {
		t.Fatalf("read multipart form: %v", err)
	}
	t.Cleanup(func() {
		_ = form.RemoveAll()
	})

	files := form.File["cover"]
	if len(files) != 1 {
		t.Fatalf("expected one file, got %d", len(files))
	}

	return files[0]
}

func useTempWorkingDirectory(t *testing.T) string {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	return tempDir
}

func TestSaveCover_OK(t *testing.T) {
	tempDir := useTempWorkingDirectory(t)

	content := []byte("fake PNG content")
	fileHeader := newFileHeader(t, "cover.PNG", content)

	path, err := SaveCover(7, fileHeader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(
		"uploads",
		"covers",
		"book_7.png",
	)

	if path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, path)
	}

	fullPath := filepath.Join(tempDir, path)

	savedContent, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read saved cover: %v", err)
	}

	if !bytes.Equal(savedContent, content) {
		t.Fatalf(
			"expected content %q, got %q",
			content,
			savedContent,
		)
	}
}

func TestSaveCover_InvalidExtension(t *testing.T) {
	fileHeader := newFileHeader(
		t,
		"cover.gif",
		[]byte("fake image"),
	)

	path, err := SaveCover(1, fileHeader)

	if !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}

	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
}

func TestSaveCover_FileTooLarge(t *testing.T) {
	fileHeader := &multipart.FileHeader{
		Filename: "cover.png",
		Size:     maxCoverSizeBytes + 1,
	}

	path, err := SaveCover(1, fileHeader)

	if !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}

	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
}

func TestSaveCover_JPEG_OK(t *testing.T) {
	useTempWorkingDirectory(t)

	fileHeader := newFileHeader(
		t,
		"photo.JPEG",
		[]byte("fake JPEG content"),
	)

	path, err := SaveCover(15, fileHeader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(
		"uploads",
		"covers",
		"book_15.jpeg",
	)

	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}
