package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// "github.com/minio/minio-go/v7"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func createBuckets(s3Client *minio.Client) {
	for cameraNum := 0; cameraNum < 10; cameraNum++ {
		bucketName := fmt.Sprintf("camera%d", cameraNum)
		ctx := context.Background()
		err := s3Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: "ru-central1"})

		if err != nil {
			// Check if bucket already exists
			exists, errBucketExists := s3Client.BucketExists(ctx, bucketName)
			if errBucketExists == nil && exists {
				fmt.Printf("Bucket %q already exists\n", bucketName)
			} else {
				log.Fatalln(err)
			}
		} else {
			fmt.Printf("Successfully created bucket %q\n", bucketName)
		}
	}
}

func saveImageToS3(s3Client *minio.Client, cameraNum int) {

	imgPath := os.Getenv("IMG_PATH")
	if imgPath == "" {
		imgPath = "./images/"
	}

	dirPath := fmt.Sprintf("%sc%d/", imgPath, cameraNum)

	// read files in directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error (dir not found):", err)
		return
	}

	for _, entry := range entries {
		filePath := fmt.Sprintf("%sc%d/%s", imgPath, cameraNum, entry.Name())
		if !strings.Contains(entry.Name(), ".lock") {
			object, err := os.Open(filePath)
			if err != nil {
				log.Fatalln(err)
			}
			defer object.Close()

			objectStat, err := object.Stat()
			if err != nil {
				log.Fatalln(err)
			}
			tags := map[string]string{
				"CameraID":       fmt.Sprintf("%d", cameraNum),
				"TimeCreation":   objectStat.ModTime().Format("2006-01-02 15:04:05.000"),
				"TimeSavingInS3": time.Now().Format("2006-01-02 15:04:05.000"),
			}

			if _, err := s3Client.PutObject(context.Background(), fmt.Sprintf("camera%d", cameraNum), fmt.Sprintf("%s_%s", objectStat.ModTime().Format("20060102150405000"), objectStat.Name()), object, objectStat.Size(), minio.PutObjectOptions{UserTags: tags}); err != nil {
				log.Fatalln(err)
			}
			log.Printf("Uploaded %s of size %d bytes to bucket %s succesfully. Modification time: %s", objectStat.Name(), objectStat.Size(), fmt.Sprintf("camera%d", cameraNum), objectStat.ModTime().Format("2006-01-02 15:04:05.000"))

			if ok := os.Remove(filepath.Join(dirPath, entry.Name())); ok != nil {
				log.Fatalln("Cannot delete file:", entry.Name())
			}
		}
	}
}

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

	createBuckets(s3Client)

	var wg sync.WaitGroup

	for cameraNum := 0; cameraNum < 10; cameraNum++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				saveImageToS3(s3Client, cameraNum)
			}

		}(cameraNum)
	}

	wg.Wait()
	fmt.Println("All goroutines have finished.")
}
