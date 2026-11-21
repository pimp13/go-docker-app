package service

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
	"test/internal/config"
)

func TestSaveImageStoresPNGUnderFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	svc := NewUploadService(config.UploadConfig{
		Dir:           dir,
		PublicPath:    "/uploads",
		MaxBytes:      1 << 20,
		PublicBaseURL: "http://localhost:8080",
	})

	file := makePNGUpload(t, "avatar.png")
	result, err := svc.SaveImage(file, "avatars")
	if err != nil {
		t.Fatalf("SaveImage: %v", err)
	}

	if result.Folder != "avatars" {
		t.Fatalf("folder = %q, want avatars", result.Folder)
	}
	if result.ContentType != "image/png" {
		t.Fatalf("content type = %q, want image/png", result.ContentType)
	}
	if result.Path == "" || result.URL == "" {
		t.Fatalf("expected path and url, got %+v", result)
	}

	stored := filepath.Join(dir, "avatars", result.Filename)
	if _, err := os.Stat(stored); err != nil {
		t.Fatalf("stored file missing: %v", err)
	}
}

func TestSaveImageRejectsInvalidFolder(t *testing.T) {
	t.Parallel()

	svc := NewUploadService(config.UploadConfig{
		Dir:           t.TempDir(),
		PublicPath:    "/uploads",
		MaxBytes:      1 << 20,
		PublicBaseURL: "http://localhost:8080",
	})

	_, err := svc.SaveImage(makePNGUpload(t, "x.png"), "../etc")
	if err != ErrUploadInvalidFolder {
		t.Fatalf("err = %v, want ErrUploadInvalidFolder", err)
	}
}

func makePNGUpload(t *testing.T, filename string) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write(buf.Bytes()); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	files := form.File["file"]
	if len(files) == 0 {
		t.Fatal("expected file part")
	}
	return files[0]
}
