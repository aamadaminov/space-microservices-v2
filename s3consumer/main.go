package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/aamadaminov/space-microservices-v2/s3consumer/adapter/s3"
	"github.com/aamadaminov/space-microservices-v2/s3consumer/config"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {

	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	s3Client, err := minio.New(cfg.Minio.MinioAddr, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.MinioUser, cfg.Minio.MinioPassword, ""),
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
				s3.SaveImageToS3(cfg, s3Client, cameraNum)
			}

		}(cameraNum)
	}

	wg.Wait()
	fmt.Println("All goroutines have finished.")
}
