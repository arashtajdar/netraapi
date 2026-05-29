package storage

import (
	"mime/multipart"
)

type Provider interface {
	UploadFile(file multipart.File, header *multipart.FileHeader) (string, error)
	DeleteFile(fileUrl string) error
}

var ActiveProvider Provider
