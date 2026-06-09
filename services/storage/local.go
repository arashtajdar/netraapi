package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalStorage struct {
	BaseDir string
	BaseURL string
}

func NewLocalStorage(baseDir, baseUrl string) *LocalStorage {
	os.MkdirAll(baseDir, os.ModePerm)
	return &LocalStorage{BaseDir: baseDir, BaseURL: baseUrl}
}

func (l *LocalStorage) UploadFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	dstPath := filepath.Join(l.BaseDir, filename)
	
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", l.BaseURL, filename), nil
}

func (l *LocalStorage) DeleteFile(fileUrl string) error {
	filename := filepath.Base(fileUrl)
	filePath := filepath.Join(l.BaseDir, filename)
	
	cleanedPath := filepath.Clean(filePath)
	expectedPrefix, err := filepath.Abs(l.BaseDir)
	if err != nil {
		return err
	}
	actualPath, err := filepath.Abs(cleanedPath)
	if err != nil {
		return err
	}
	
	if !strings.HasPrefix(actualPath, expectedPrefix) {
		return fmt.Errorf("path traversal attempt detected")
	}

	return os.Remove(actualPath)
}
