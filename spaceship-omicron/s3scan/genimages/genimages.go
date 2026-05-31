package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func mkDirs() {

	imgPath := os.Getenv("IMG_PATH")
	if imgPath == "" {
		imgPath = "./images/"
	}
	for cameraNum := 0; cameraNum < 10; cameraNum++ {
		os.Mkdir(fmt.Sprintf("%sc%d", imgPath, cameraNum), 0777)
	}

}

func randomImageAdd(cameraNum int, pauseTime time.Duration) {

	dirPathSource := "./sourceimages/"

	imgPath := os.Getenv("IMG_PATH")
	if imgPath == "" {
		imgPath = "./images/"
	}

	dirPathForSave := fmt.Sprintf("%sc%d/", imgPath, cameraNum)

	for {
		fileCnt := 0
		for fileCnt = 0; fileCnt < 50; fileCnt++ {
			n := rand.IntN(10)
			imageSource, err := os.ReadFile(fmt.Sprintf("%s%d.jpg", dirPathSource, n))
			if err != nil {
				fmt.Println("Error file reading:", err)
				return
			}
			fileNameLocked := fmt.Sprintf("DSC%04d.jpg.lock", fileCnt)
			fileName := fmt.Sprintf("DSC%04d.jpg", fileCnt)
			os.WriteFile(filepath.Join(dirPathForSave, fileNameLocked), imageSource, 0644)
			fmt.Printf("Image from Camera %d received succesfully. Time: %s\n", cameraNum, time.Now().Format("2006-01-02 15:04:05.000"))
			os.Rename(filepath.Join(dirPathForSave, fileNameLocked), filepath.Join(dirPathForSave, fileName))
			time.Sleep(pauseTime * time.Millisecond)
		}
	}
}

func main() {
	var wg sync.WaitGroup
	var pauseTime time.Duration = 5000

	mkDirs()

	for cameraNum := 0; cameraNum < 10; cameraNum++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			randomImageAdd(cameraNum, pauseTime)
		}(cameraNum)
	}
	wg.Wait() // Block until the WaitGroup counter becomes zero
	fmt.Println("All goroutines have finished.")

}
