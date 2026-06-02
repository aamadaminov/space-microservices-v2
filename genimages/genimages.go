package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aamadaminov/space-microservices-v2/genimages/config"
	"github.com/aamadaminov/space-microservices-v2/genimages/config/paths"
)

func mkDirs(cfg paths.Config) error {

	os.Mkdir(fmt.Sprintf("%s", cfg.ImgPath), 0777)
	for cameraNum := 0; cameraNum < 10; cameraNum++ {
		os.Mkdir(fmt.Sprintf("%s/c%d", cfg.ImgPath, cameraNum), 0777)
	}
	return nil
}

func randomImageAdd(cfg paths.Config, cameraNum int, pauseTime time.Duration) error {
	dirPathForSave := fmt.Sprintf("%s/c%d/", cfg.ImgPath, cameraNum)
	for {
		fileCnt := 0
		for fileCnt = 0; fileCnt < 50; fileCnt++ {
			n := rand.IntN(10)
			imageSource, err := os.ReadFile(fmt.Sprintf("%s/%d.jpg", cfg.DirPathSource, n))
			if err != nil {
				fmt.Println("Error file reading:", err)
				return err
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

	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err)
	}
	mkDirs(cfg.Paths)

	wg := &sync.WaitGroup{}
	var pauseTime time.Duration = 5000

	for cameraNum := 0; cameraNum < 10; cameraNum++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			randomImageAdd(cfg.Paths, cameraNum, pauseTime)
		}(cameraNum)
	}
	wg.Wait()
	fmt.Println("All goroutines have finished.")
}
