package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"test/internal/config"
	"test/internal/dto"

	"github.com/google/uuid"
)

var (
	ErrUploadMissingFile   = errors.New("file is required")
	ErrUploadTooLarge      = errors.New("file exceeds the maximum allowed size")
	ErrUploadInvalidType   = errors.New("unsupported file type")
	ErrUploadInvalidFolder = errors.New("invalid upload folder")
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var allowedFolders = map[string]struct{}{
	"avatars":  {},
	"articles": {},
	"general":  {},
}

// UploadService stores uploaded files and returns public URLs.
type UploadService interface {
	SaveImage(file *multipart.FileHeader, folder string) (dto.UploadResponse, error)
}

type uploadService struct {
	cfg config.UploadConfig
}

// NewUploadService creates an upload service rooted at cfg.Dir.
func NewUploadService(cfg config.UploadConfig) UploadService {
	_ = os.MkdirAll(cfg.Dir, 0o755)
	return &uploadService{cfg: cfg}
}

func (s *uploadService) SaveImage(file *multipart.FileHeader, folder string) (dto.UploadResponse, error) {
	if file == nil {
		return dto.UploadResponse{}, ErrUploadMissingFile
	}
	if file.Size > s.cfg.MaxBytes {
		return dto.UploadResponse{}, ErrUploadTooLarge
	}

	folder, err := normalizeFolder(folder)
	if err != nil {
		return dto.UploadResponse{}, err
	}

	src, err := file.Open()
	if err != nil {
		return dto.UploadResponse{}, fmt.Errorf("open upload: %w", err)
	}
	defer src.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return dto.UploadResponse{}, fmt.Errorf("read upload: %w", err)
	}
	head = head[:n]

	contentType := http.DetectContentType(head)
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return dto.UploadResponse{}, ErrUploadInvalidType
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return dto.UploadResponse{}, fmt.Errorf("rewind upload: %w", err)
	}

	filename := uuid.NewString() + ext
	dir := filepath.Join(s.cfg.Dir, folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return dto.UploadResponse{}, fmt.Errorf("create upload dir: %w", err)
	}

	destPath := filepath.Join(dir, filename)
	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return dto.UploadResponse{}, fmt.Errorf("create file: %w", err)
	}
	defer dest.Close()

	written, err := io.Copy(dest, io.LimitReader(src, s.cfg.MaxBytes+1))
	if err != nil {
		_ = os.Remove(destPath)
		return dto.UploadResponse{}, fmt.Errorf("write upload: %w", err)
	}
	if written > s.cfg.MaxBytes {
		_ = os.Remove(destPath)
		return dto.UploadResponse{}, ErrUploadTooLarge
	}

	publicPath := strings.TrimRight(s.cfg.PublicPath, "/") + "/" + folder + "/" + filename
	base := strings.TrimRight(s.cfg.PublicBaseURL, "/")

	return dto.UploadResponse{
		URL:         base + publicPath,
		Path:        publicPath,
		Filename:    filename,
		Size:        written,
		ContentType: contentType,
		Folder:      folder,
	}, nil
}

func normalizeFolder(folder string) (string, error) {
	folder = strings.ToLower(strings.TrimSpace(folder))
	if folder == "" {
		folder = "general"
	}
	if _, ok := allowedFolders[folder]; !ok {
		return "", ErrUploadInvalidFolder
	}
	return folder, nil
}
