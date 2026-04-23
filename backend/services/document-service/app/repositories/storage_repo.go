package repositories

import (
	"context"
	"mime/multipart"
	"os"
	"strings"
	"fmt"

	"github.com/minio/minio-go/v7"
)

type StorageRepo struct{
	minio *minio.Client
}

func NewStorageRepo(minio *minio.Client) *StorageRepo{
	return &StorageRepo{
		minio : minio,
	}
}

func (s *StorageRepo) Upload_document(file multipart.File, header *multipart.FileHeader, filename string)(string, string, error){
	fileName := strings.Join(strings.Fields(filename), "_")

	_, err := s.minio.PutObject(
		context.Background(),
		"testing",
		fileName,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: header.Header.Get("Content-Type"),
			ContentDisposition: "inline",
		},
	)
	if err != nil{
		return "", "", err
	}

	filepath := fmt.Sprintf(os.Getenv("BUCKET_URL"), filename)
	return filename, filepath, nil
}

func (s *StorageRepo) Delete_document(filepath string) error {
	 _, err := s.minio.StatObject(
		context.Background(),
		"testing",
		filepath,
		minio.StatObjectOptions{},
	)
    if err != nil {
        return err 
   }

	err = s.minio.RemoveObject(
		context.Background(), 
		"testing",
		filepath,
		minio.RemoveObjectOptions{},
	)
	if err != nil{
		return err
	}

	return nil
}