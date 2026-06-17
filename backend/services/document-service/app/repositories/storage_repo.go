package repositories

import (
	"context"
	"mime/multipart"
	"os"
	"io"
	"strings"
	"fmt"
	"log"
	"path/filepath"
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

var privateBucket = os.Getenv("PRIVATE_BUCKET")
var publicBucket = os.Getenv("PUBLIC_BUCKET")

func (s *StorageRepo) Upload_document(file io.Reader, header *multipart.FileHeader, filename string,isPublic bool)(string, string, error){
	extension := filepath.Ext(header.Filename)
	fileName := strings.Join(strings.Fields(filename), "_") + extension

	var err error
	var filepath string
	if isPublic{
		_, err = s.minio.PutObject(
			context.Background(),
			publicBucket,
			fileName,
			file,
			header.Size,
			minio.PutObjectOptions{
				ContentType: header.Header.Get("Content-Type"),
				ContentDisposition: "inline",
			},
		)
		filepath = fmt.Sprintf(os.Getenv("PUBLIC_BUCKET_URL"), filename + extension)
	}else{
		_, err = s.minio.PutObject(
			context.Background(),
			privateBucket,
			fileName,
			file,
			header.Size,
			minio.PutObjectOptions{
				ContentType: header.Header.Get("Content-Type"),
				ContentDisposition: "inline",
			},
		)
		filepath = fmt.Sprintf(os.Getenv("PRIVATE_BUCKET_URL"), filename + extension)
	}

	if err != nil{
		return "", "", err
	}

	return fileName, filepath, nil
}

func (s *StorageRepo) Update_document(file multipart.File, header *multipart.FileHeader, oldFilePath, newFileName string) (string, string, error) {
    bucket, _ := s.resolveBucket(oldFilePath)
    extension := filepath.Ext(header.Filename)
    fileName := strings.Join(strings.Fields(newFileName), "_") + extension

    _, err := s.minio.PutObject(context.Background(), bucket, fileName, file, header.Size,
        minio.PutObjectOptions{
            ContentType:        header.Header.Get("Content-Type"),
            ContentDisposition: "inline",
        },
    )
    if err != nil {
        return "", "", err
    }

    var newFilePath string
    if bucket == publicBucket {
        newFilePath = fmt.Sprintf(os.Getenv("PUBLIC_BUCKET_URL"), fileName)
    } else {
        newFilePath = fmt.Sprintf(os.Getenv("PRIVATE_BUCKET_URL"), fileName)
    }
    return fileName, newFilePath, nil
}

func (s *StorageRepo) Update_documentName(oldFilePath, newFileName string) (string, string, error) {
    bucket, objectName := s.resolveBucket(oldFilePath)
    extension := filepath.Ext(objectName)
    newObjectName := strings.Join(strings.Fields(newFileName), "_") + extension

    _, err := s.minio.CopyObject(context.Background(),
        minio.CopyDestOptions{Bucket: bucket, Object: newObjectName},
        minio.CopySrcOptions{Bucket: bucket, Object: objectName},
    )
    if err != nil {
        return "", "", err
    }

    err = s.minio.RemoveObject(context.Background(), bucket, objectName, minio.RemoveObjectOptions{})
    if err != nil {
        log.Print("Error deleting old file", err)
    }

    var newFilePath string
    if bucket == publicBucket {
        newFilePath = fmt.Sprintf(os.Getenv("PUBLIC_BUCKET_URL"), newObjectName)
    } else {
        newFilePath = fmt.Sprintf(os.Getenv("PRIVATE_BUCKET_URL"), newObjectName)
    }
    return newObjectName, newFilePath, nil
}

func (s *StorageRepo) Update_documentFile(file multipart.File, header *multipart.FileHeader, oldFilePath string) (string, string, error) {
    bucket, objectName := s.resolveBucket(oldFilePath)

    _, err := s.minio.PutObject(context.Background(), bucket, objectName, file, header.Size,
        minio.PutObjectOptions{
            ContentType:        header.Header.Get("Content-Type"),
            ContentDisposition: "inline",
        },
    )
    if err != nil {
        return "", "", err
    }
	log.Print("old ", oldFilePath)
	log.Print("object name", objectName)

    return objectName, oldFilePath, nil
}


func (s *StorageRepo) Delete_document(filepath string) error {
	bucket, objectName := s.resolveBucket(filepath)
   
	err := s.minio.RemoveObject(
		context.Background(), 
		bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
	if err != nil{
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



/* generate presign (cannot without public domain alamak) || can if the application ran directly without container
func (s *StorageRepo) Get_document_object(filename string) (string, error){
	fileName := strings.Join(strings.Fields(filename), "_")
	expiry := 1*4*time.Hour

	filepath, err := s.minio.PresignedGetObject(
		context.Background(),
		"testing2",
		fileName,
		expiry,
		nil,
	)
	if err != nil {
		return "", err
	}
	
	return filepath.String(), nil
}*/