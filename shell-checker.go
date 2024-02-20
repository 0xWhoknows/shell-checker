// go run check.go -f a.txt
// want to run it globally ? build it first : go build check.go && chmod +x check && mv /home/<your-linux-user>/go/bin

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	shellTitle  string
	shellFile   string
	retryCount  int
	userAgent   string
)

func init() {
	flag.StringVar(&shellTitle, "t", "<title>403WebShell</title>", "Site title") // you can add here ur shell title then commend will be go run check.go -f a.txt
	flag.StringVar(&shellFile, "f", "", "The list of shells")
	flag.IntVar(&retryCount, "r", 1, "Retrie number")
	flag.StringVar(&userAgent, "ua", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.95 Safari/537.36", "Specify the User-Agent header")
	flag.Parse()

	if shellTitle == "" || shellFile == "" {
		log.Fatalf("Usage: file.go -t <site-title> -f <shell-file> [-r <retry-count>] [-ua <user-agent>]")
		os.Exit(1)
	}
}

func clear() {
	var clearCmd string

	if system := strings.ToLower(os.Getenv("GOOS")); system == "windows" {
		clearCmd = "cls"
	} else {
		clearCmd = "clear"
	}

	cmd := exec.Command(clearCmd)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func checkHost(host string, wg *sync.WaitGroup, resultCh chan<- string, semaphore chan struct{}) {
	defer wg.Done()
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Ignore SSL verify
		},
	}

	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		log.Printf("Error creating request: %s", err)
		resultCh <- fmt.Sprintf("Error creating request: %s", host)
		return
	}
	req.Header.Set("User-Agent", userAgent)

	for i := 0; i < retryCount; i++ {
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			if strings.Contains(string(body), shellTitle) {
				log.Printf("[XxX] %s => Shell Alive!\n", host)
				resultCh <- host + " => Shell Alive!"
				return
			} else {
				log.Printf("Shell Dead! %s\n", host)
				resultCh <- host + " => Shell Dead!"
				return
			}
		}
		time.Sleep(1 * time.Second) //  retrying
	}

	log.Printf("Failed  %d retries! %s\n", retryCount, host)
	resultCh <- fmt.Sprintf("Failed  %d retries! %s", retryCount, host)
}

func main() {
	clear()

	var wg sync.WaitGroup
	resultCh := make(chan string)
	semaphore := make(chan struct{}, 10)

	hosts, err := ioutil.ReadFile(shellFile)
	if err != nil {
		log.Fatalf("Error reading file: %s", err)
		return
	}
	hostList := strings.Split(string(hosts), "\n")

	for _, host := range hostList {
		if host != "" {
			wg.Add(1)
			go checkHost(host, &wg, resultCh, semaphore)
		}
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	aliveFilePath := filepath.Join(".", "Alivexhell.txt")
	deadFilePath := filepath.Join(".", "deadxhell.txt")
	notWorksFilePath := filepath.Join(".", "notworks.txt")

	aliveFile, err := os.OpenFile(aliveFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening alive file: %s", err)
		return
	}
	defer aliveFile.Close()

	deadFile, err := os.OpenFile(deadFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening dead file: %s", err)
		return
	}
	defer deadFile.Close()

	notWorksFile, err := os.OpenFile(notWorksFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening not works file: %s", err)
		return
	}
	defer notWorksFile.Close()

	for result := range resultCh {
		switch {
		case strings.Contains(result, "Shell Alive"):
			hostURL := strings.Split(result, " => ")[0]
			aliveFile.WriteString(hostURL + "\n")
		case strings.Contains(result, "Shell Dead"):
			hostURL := strings.Split(result, " => ")[0]
			deadFile.WriteString(hostURL + "\n")
		default:
			notWorksFile.WriteString(result + "\n")
		}
	}
}