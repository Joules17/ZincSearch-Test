package main

// imports
import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Email Object Struct
type Email struct {
	MessageID               string `json:"message_id"`
	Date                    string `json:"date"`
	From                    string `json:"from"`
	To                      string `json:"to"`
	Subject                 string `json:"subject"`
	MimeVersion             string `json:"mime_version"`
	ContentType             string `json:"content_type"`
	ContentTransferEncoding string `json:"content_transfer_encoding"`
	XFrom                   string `json:"x_from"`
	XTo                     string `json:"x_to"`
	Xcc                     string `json:"x_cc"`
	Xbcc                    string `json:"x_bcc"`
	XFolder                 string `json:"x_folder"`
	XOrigin                 string `json:"x_origin"`
	XFileName               string `json:"x_file_name"`
	Content                 string `json:"content"`
}

// EmailPkg Struct
type EmailPkg struct {
	Emails []Email `json:"emails"`
}

// constant settings
const (
	emailsPerPkg   = 1000
	workers        = 6
	ZincSearchURL  = "http://localhost:4080/api/_bulkv2"
	ZincSearchUser = "admin"
	ZincSearchPass = "Complexpass#123"
)

// wait groups
var mainWg sync.WaitGroup

// func sendToZincSearch
// based on email Pkg provided sends a bulkv2 petition to ZincSearch
func sendToZincSearch(emails []Email) {
	fmt.Println("Sending package to ZincSearch...")
	reqBody, err := json.Marshal(map[string]interface{}{
		"index":   "emails",
		"records": emails,
	})
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// fmt.Println("Request Body Sending: ", string(reqBody))

	req, err := http.NewRequest("POST", ZincSearchURL, strings.NewReader(string(reqBody)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.SetBasicAuth(ZincSearchUser, ZincSearchPass)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	client := &http.Client{
		Timeout: time.Second * 10, // timeout of 10 sec
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error in response. Status code:", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Package Sent successfully!")
	fmt.Println("Response Body:", string(body))
}

// func worker
// worker: process pkgs and send to zincsearch
func worker(pkgChan <-chan EmailPkg, wg *sync.WaitGroup) {
	defer wg.Done()

	var emails []Email

	for pkg := range pkgChan {
		emails = append(emails, pkg.Emails...)

		if len(emails) >= emailsPerPkg {
			sendToZincSearch(emails[:emailsPerPkg])
			emails = emails[emailsPerPkg:]
		}
	}

	// send other emails after closing the channel
	if len(emails) > 0 {
		sendToZincSearch(emails)
	}
}

// readDirectory function
// look for files given a path as arg (recursively)
func readDirectory(directoryPath string, wg *sync.WaitGroup, pkgChan chan<- EmailPkg) {
	defer wg.Done()

	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		log.Fatal(err)
	}

	var emails []Email

	for _, file := range files {
		filePath := filepath.Join(directoryPath, file.Name())

		if file.IsDir() {
			wg.Add(1)
			go readDirectory(filePath, wg, pkgChan)
		} else {
			fileContent, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Printf("Error reading file %s: %v", filePath, err)
				continue
			}

			email, err := scanFile(file.Name(), string(fileContent))
			if err != nil {
				log.Printf("Error scanning file %s: %v", filePath, err)
				continue
			}

			emails = append(emails, email)

			if len(emails) >= emailsPerPkg {
				pkgChan <- EmailPkg{Emails: emails[:emailsPerPkg]}
				emails = emails[emailsPerPkg:]
			}
		}
	}

	if len(emails) > 0 {
		pkgChan <- EmailPkg{Emails: emails}
	}
}

// scanFile Function
// scan file content and create a new email object
func scanFile(fileName, fileContent string) (Email, error) {
	// email struct
	email := Email{}

	// Content Lines
	var contentLines []string

	// scanner
	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Message-ID: ") {
			// found MessageID but parse it a little bit more
			idStart := strings.Index(line, "<") + 1
			idEnd := strings.Index(line, ">")
			if idStart != -1 && idEnd != -1 && idEnd > idStart {
				messageID := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line[idStart:idEnd], ".JavaMail.evans@thyme"), ".JavaMail.evans@thyme"))
				if messageID == "" {
					// Null file
					return Email{}, fmt.Errorf("Invalid MessageID in file %s", fileName)
				}
				email.MessageID = messageID
			}
		} else if strings.HasPrefix(line, "Date: ") {
			email.Date = strings.TrimSpace(strings.TrimPrefix(line, "Date: "))
		} else if strings.HasPrefix(line, "From: ") {
			email.From = strings.TrimSpace(strings.TrimPrefix(line, "From: "))
		} else if strings.HasPrefix(line, "To: ") {
			email.To = strings.TrimSpace(strings.TrimPrefix(line, "To: "))
		} else if strings.HasPrefix(line, "Subject: ") {
			email.Subject = strings.TrimSpace(strings.TrimPrefix(line, "Subject: "))
		} else if strings.HasPrefix(line, "Mime-Version: ") {
			email.MimeVersion = strings.TrimSpace(strings.TrimPrefix(line, "Mime-Version: "))
		} else if strings.HasPrefix(line, "Content-Type: ") {
			email.ContentType = strings.TrimSpace(strings.TrimPrefix(line, "Content-Type: "))
		} else if strings.HasPrefix(line, "Content-Transfer-Encoding: ") {
			email.ContentTransferEncoding = strings.TrimSpace(strings.TrimPrefix(line, "Content-Transfer-Encoding: "))
		} else if strings.HasPrefix(line, "X-From: ") {
			email.XFrom = strings.TrimSpace(strings.TrimPrefix(line, "X-From: "))
		} else if strings.HasPrefix(line, "X-To: ") {
			email.XTo = strings.TrimSpace(strings.TrimPrefix(line, "X-To: "))
		} else if strings.HasPrefix(line, "X-cc: ") {
			email.Xcc = strings.TrimSpace(strings.TrimPrefix(line, "X-cc: "))
		} else if strings.HasPrefix(line, "X-bcc: ") {
			email.Xbcc = strings.TrimSpace(strings.TrimPrefix(line, "X-bcc: "))
		} else if strings.HasPrefix(line, "X-Folder: ") {
			email.XFolder = strings.TrimSpace(strings.TrimPrefix(line, "X-Folder: "))
		} else if strings.HasPrefix(line, "X-Origin: ") {
			email.XOrigin = strings.TrimSpace(strings.TrimPrefix(line, "X-Origin: "))
		} else if strings.HasPrefix(line, "X-FileName: ") {
			email.XFileName = strings.TrimSpace(strings.TrimPrefix(line, "X-FileName: "))
		} else {
			// Content
			contentLines = append(contentLines, line)
		}
	}
	// null file?
	if email.MessageID == "" {
		return Email{}, fmt.Errorf("No MessageID in file %s", fileName)
	}

	// email content
	email.Content = strings.Join(contentLines, "\n")

	return email, nil
}

// main: main function of indexer program
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file to index.")
		os.Exit(1)
	}

	mainWg.Add(1)
	directoryPath := os.Args[1]

	fmt.Println("Main goroutine started")
	defer fmt.Println("Main goroutine completed")

	pkgChan := make(chan EmailPkg, 500)

	// initialize workers
	var workerWg sync.WaitGroup
	for i := 0; i < workers; i++ {
		workerWg.Add(1)
		go worker(pkgChan, &workerWg)
	}

	// reading dirs
	var readDirWg sync.WaitGroup
	readDirWg.Add(1)
	go readDirectory(directoryPath, &readDirWg, pkgChan)

	// closing the channel when all the work is done
	go func() {
		readDirWg.Wait()
		close(pkgChan)
		workerWg.Wait()
		mainWg.Done()
	}()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		log.Println(http.ListenAndServe("localhost:6060", mux))
	}()

	// counting time of execution
	startTime := time.Now()

	// wait for all goroutines to finish
	mainWg.Wait()

	// time elapsed
	elapsedTime := time.Since(startTime)
	fmt.Printf("Processing took %s\n", elapsedTime)
}
