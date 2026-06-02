package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aamadaminov/space-microservices-v2/s3consumer/adapter/s3"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {

	minioAddr := os.Getenv("MINIO_ENDPOINT")
	if minioAddr == "" {
		minioAddr = "localhost:9050"
	}
	minioUser := os.Getenv("MINIO_USER")
	if minioUser == "" {
		minioUser = "minioadmin"
	}
	minioPassword := os.Getenv("MINIO_PASSWORD")
	if minioPassword == "" {
		minioPassword = "minioadmin"
	}
	s3Client, err := minio.New(minioAddr, &minio.Options{
		Creds:  credentials.NewStaticV4(minioUser, minioPassword, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	s3.CreateBuckets(s3Client)

	var wg sync.WaitGroup

	for cameraNum := 0; cameraNum < 10; cameraNum++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				s3.SaveImageToS3(s3Client, cameraNum)
			}

		}(cameraNum)
	}

	wg.Wait()
	fmt.Println("All goroutines have finished.")
}
