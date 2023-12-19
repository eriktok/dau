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

var (
	outputDir = flag.String("o", "js", "output directory name")
	inputFile = flag.String("i", "url_list.txt", "input file name")
)

// User-Agent header to mimic a browser
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"

var downloadCount int32
var mutex sync.Mutex

func main() {
	flag.Parse()

	currentDir, err := os.Getwd()
	joined := filepath.Join(currentDir, *outputDir)
	path := joined + "/"
	createFolderIfNotExist(joined)
	if err != nil {
		log.Fatalf("Could not get the current working directory: %v", err)
	}
	fmt.Println("Current directory:", currentDir)

	// Open the file containing URLs
	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	urls := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from the file: %v", err)
	}

	downloadCount = 0
	downloadUrls(urls, path)

	fmt.Println("Total JavaScript files downloaded:", downloadCount)
}

func downloadUrls(urls []string, path string) {
	var wg sync.WaitGroup
	client := &http.Client{}

	wg.Add(len(urls))

	for _, url := range urls {
		go func(url string) {
			defer wg.Done()
			tokens := strings.Split(url, "/")
			fileName := tokens[len(tokens)-1]
			fmt.Println("Downloading", url, "to", fileName)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Println("Error creating request for", url, "-", err)
				return
			}

			req.Header.Set("User-Agent", userAgent)

			res, err := client.Do(req)
			if err != nil {
				log.Println("HTTP GET error:", err)
				return
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				log.Println("Unexpected status code:", res.StatusCode)
				return
			}

			// Check if the file already exists in the destination folder
			if !fileExists(path + fileName) {
				output, err := os.Create(path + fileName)
				if err != nil {
					log.Println("Error while creating", fileName, "-", err)
					return
				}
				defer output.Close()

				// Copy the entire JS content to the file
				_, err = io.Copy(output, res.Body)
				if err != nil {
					log.Println("Error while downloading", url, "-", err)
					return
				}

				fmt.Println("Downloaded", fileName)
				atomicAdd()
			} else {
				fmt.Println("File", fileName, "already exists in the destination folder. Skipping.")
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
			log.Println(err)
		}
		fmt.Println("Folder created:", folderPath)
	} else {
		fmt.Println("Folder already exists:", folderPath)
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// This is git test.
func atomicAdd() {
	mutex.Lock()
	defer mutex.Unlock()
	downloadCount++
}
