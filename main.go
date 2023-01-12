package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var outputDir = flag.String("o", "/js", " output directory name")

func main() {
	flag.Parse()
	currentDir, err := os.Getwd()
	joined := filepath.Join(currentDir, *outputDir)
	path := joined + "/"
	createFolderIfNotExist(joined)
	if err != nil {
		log.Fatalf("Could not get current working directory: %v", err)
	}
	fmt.Println(currentDir)
	urls := make([]string, 0)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from stdin: %v", err)
	}
	downloadUrls(urls, path)
}

func downloadUrls(urls []string, path string) {
	var wg sync.WaitGroup

	wg.Add(len(urls))

	for _, url := range urls {
		go func(url string) {
			defer wg.Done()
			tokens := strings.Split(url, "/")
			fileName := tokens[len(tokens)-1]
			fmt.Println("Downloading", url, "to", fileName)

			output, err := os.Create(path + fileName)
			if err != nil {
				log.Fatal("Error while creating", fileName, "-", err)
			}
			defer output.Close()

			res, err := http.Get(url)
			if err != nil {
				log.Fatal("http get error: ", err)
			} else {
				defer res.Body.Close()
				_, err = io.Copy(output, res.Body)
				if err != nil {
					log.Fatal("Error while downloading", url, "-", err)
				} else {
					fmt.Println("Downloaded", fileName)
				}
			}
		}(url)
	}
	wg.Wait()
	fmt.Println("Done")
}

func createFolderIfNotExist(folderPath string) {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Folder created: ", folderPath)
	} else {
		fmt.Println("Folder already exist: ", folderPath)
	}
}
