package repositories

import (
	"capstone/app/core/utils"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type StorageRepo struct {
	minio *minio.Client
}

func NewStorageRepo(minio *minio.Client) *StorageRepo {
	return &StorageRepo{
		minio: minio,
	}
}

var privateBucket = os.Getenv("PRIVATE_BUCKET")
var publicBucket = os.Getenv("PUBLIC_BUCKET")

func (s *StorageRepo) Upload_document(file io.Reader, objectId string, isPublic bool, size int64, contentType string) (string, error) {
	var err error
	var filepath string
	if isPublic {
		_, err = s.minio.PutObject(
			context.Background(),
			publicBucket,
			objectId,
			file,
			size,
			minio.PutObjectOptions{
				ContentType:        contentType,
				ContentDisposition: "inline",
			},
		)
		filepath = fmt.Sprintf(os.Getenv("PUBLIC_BUCKET_URL"), objectId)
	} else {
		_, err = s.minio.PutObject(
			context.Background(),
			privateBucket,
			objectId,
			file,
			size,
			minio.PutObjectOptions{
				ContentType:        contentType,
				ContentDisposition: "inline",
			},
		)
		filepath = fmt.Sprintf(os.Getenv("PRIVATE_BUCKET_URL"), objectId)
	}

	if err != nil {
		return "", err
	}

	return filepath, nil
}

func (s *StorageRepo) Update_document(file io.Reader, fileSize int64, contentType string, oldFilePath, newFileName, objectId string) (string, string, string, error) {
	bucket, _ := s.resolveBucket(oldFilePath)

	_, err := s.minio.PutObject(context.Background(), bucket, objectId, file, fileSize,
		minio.PutObjectOptions{
			ContentType:        contentType,
			ContentDisposition: "inline",
		},
	)
	if err != nil {
		return "", "", "", err
	}

	var newFilePath string
	if bucket == publicBucket {
		newFilePath = fmt.Sprintf(os.Getenv("PUBLIC_BUCKET_URL"), objectId)
	} else {
		newFilePath = fmt.Sprintf(os.Getenv("PRIVATE_BUCKET_URL"), objectId)
	}

	updatedHash, _ := s.GenerateObjectHMAC(objectId, oldFilePath)
	return newFileName, updatedHash, newFilePath, nil
}

func (s *StorageRepo) Update_documentFile(file io.Reader, fileSize int64, contentType string, oldFilePath, objectId string) (string, string, string, error) {
	bucket, objectName := s.resolveBucket(oldFilePath)

	_, err := s.minio.PutObject(context.Background(), bucket, objectId, file, fileSize,
		minio.PutObjectOptions{
			ContentType:        contentType,
			ContentDisposition: "inline",
		},
	)
	if err != nil {
		return "", "", "", err
	}

	updatedHash, _ := s.GenerateObjectHMAC(objectId, oldFilePath)
	return objectName, updatedHash, oldFilePath, nil
}

func (s *StorageRepo) Delete_document(filepath string) error {
	bucket, objectName := s.resolveBucket(filepath)

	err := s.minio.RemoveObject(
		context.Background(),
		bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *StorageRepo) GetMinioObject(ctx context.Context, minioPath string) (*minio.Object, *minio.ObjectInfo, error) {
	bucket, objectName := s.resolveBucket(minioPath)
	object, err := s.minio.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}

	stat, err := object.Stat()
	if err != nil {
		return nil, nil, err
	}

	return object, &stat, nil
}

func (s *StorageRepo) GenerateObjectHMAC(objectId string, minioPath string) (string, error) {
	bucket, objectName := s.resolveBucket(minioPath)
	object, err := s.minio.GetObject(
		context.Background(),
		bucket,
		objectName,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return "", err
	}
	defer object.Close()

	fileBytes, _ := io.ReadAll(object)
	hash := utils.GenerateHMAC(fileBytes)

	return hash, nil
}

func (s *StorageRepo) bucket(isPublic bool) string {
	if isPublic {
		return publicBucket
	}
	return privateBucket
}

func (s *StorageRepo) resolveBucket(filepath string) (string, string) {
	publicBucketPrefix := strings.TrimSuffix(os.Getenv("PUBLIC_BUCKET_URL"), "%s")
	privateBucketPrefix := strings.TrimSuffix(os.Getenv("PRIVATE_BUCKET_URL"), "%s")
	if strings.HasPrefix(filepath, publicBucketPrefix) {
		objectName := strings.TrimPrefix(filepath, publicBucketPrefix)
		return publicBucket, objectName
	}
	objectName := strings.TrimPrefix(filepath, privateBucketPrefix)

	return privateBucket, objectName
}

func (s *StorageRepo) Get_document_presign(objectId string, isPublic bool) (string, string, error) {
	bucket := s.bucket(isPublic)
	expiry := 5 * time.Hour
	url, err := s.minio.PresignedGetObject(
		context.Background(),
		bucket,
		objectId,
		expiry,
		nil,
	)
	if err != nil {
		return "Error getting presign", "", err
	}

	contentType := ""
	info, statErr := s.minio.StatObject(context.Background(), bucket, objectId, minio.StatObjectOptions{})
	if statErr != nil {
		log.Print("Error getting object stat for content type: ", statErr)
	} else {
		contentType = info.ContentType
	}

	return url.String(), contentType, nil
}
