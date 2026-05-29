package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	Client    *s3.Client
	Bucket    string
	PublicURL string
}

func NewR2Storage(accountId, accessKeyId, secretKey, bucket, publicUrl string) *R2Storage {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		panic("Failed to load R2 config: " + err.Error())
	}
	
	client := s3.NewFromConfig(cfg)
	return &R2Storage{Client: client, Bucket: bucket, PublicURL: publicUrl}
}

func (r *R2Storage) UploadFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
    
	_, err := r.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(r.Bucket),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", err
	}
    
	return fmt.Sprintf("%s/%s", r.PublicURL, filename), nil
}

func (r *R2Storage) DeleteFile(fileUrl string) error {
	return nil
}
