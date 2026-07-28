package objectStorage

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinio() (*minio.Client, error) {
	endpoint := os.Getenv("ENDPOINT")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatal("error: ", err)
	}

	log.Println("Minio client Init")
	return minioClient, nil
}


func CreateBuckets(client *minio.Client) error {
	ctx := context.Background()

	buckets := []string{"publicbucket", "privatebucket"}

	for _, bucket := range buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return err
		}
		if exists {
			log.Printf("bucket %s already exists, skipping\n", bucket)
			continue
		}

		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
		log.Printf("bucket %s created\n", bucket)
	}

	return nil
}