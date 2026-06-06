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

var bucket = os.Getenv("BUCKET")

func (s *StorageRepo) Upload_document(file io.Reader, header *multipart.FileHeader, filename string)(string, string, error){
	extension := filepath.Ext(header.Filename)
	fileName := strings.Join(strings.Fields(filename), "_") + extension

	_, err := s.minio.PutObject(
		context.Background(),
		bucket,
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
	

	filepath := fmt.Sprintf(os.Getenv("BUCKET_URL"), fileName)
	return fileName, filepath, nil
}

func (s *StorageRepo) Update_document(file multipart.File, header *multipart.FileHeader, newFileName string) (string,string,error){
	extension := filepath.Ext(header.Filename)
	fileName := strings.Join(strings.Fields(newFileName), "_") + extension
	_, err := s.minio.PutObject(
		context.Background(),
		bucket,
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
	

	filepath := fmt.Sprintf(os.Getenv("BUCKET_URL"), fileName)
	return fileName, filepath, nil
}

func (s *StorageRepo) Update_documentName(oldFileName, newFileName string)(string,string, error){
	extension := filepath.Ext(oldFileName)
	_, err := s.minio.CopyObject(context.Background(), 
			  minio.CopyDestOptions{
				Bucket: bucket,
				Object: newFileName + extension,
			  },
			  minio.CopySrcOptions{
				Bucket: bucket,
				Object: oldFileName,
			  })
	if err != nil {
		return "","",err
	}

	err = s.minio.RemoveObject(context.Background(), bucket, oldFileName, minio.RemoveObjectOptions{})
	if err != nil{
		log.Print("Error deleting old file")
	}

	filePath := fmt.Sprintf(os.Getenv("BUCKET_URL"), newFileName + extension)
	return newFileName + extension,filePath,nil
}

func (s *StorageRepo) Update_documentFile(file multipart.File, header *multipart.FileHeader, oldFileName string)(string,string,error){
	_, err := s.minio.PutObject(
		context.Background(),
		bucket,
		oldFileName,
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
	

	filepath := fmt.Sprintf(os.Getenv("BUCKET_URL"), oldFileName)
	return oldFileName, filepath, nil
}


func (s *StorageRepo) Delete_document(filename string) error {
	fileName := strings.Join(strings.Fields(filename), "_")
	 _, err := s.minio.StatObject(
		context.Background(),
		bucket,
		fileName,
		minio.StatObjectOptions{},
	)
    if err != nil {
		log.Print("file is not exist", err)
        return err 
   }

   
	err = s.minio.RemoveObject(
		context.Background(), 
		bucket,
		fileName,
		minio.RemoveObjectOptions{},
	)
	if err != nil{
		return err
	}

	return nil
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