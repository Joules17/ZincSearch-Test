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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Email Object Struct
type Email struct {
	MessageID               string
	Date                    string
	From                    string
	To                      string
	Subject                 string
	MimeVersion             string
	ContentType             string
	ContentTransferEncoding string
	XFrom                   string
	XTo                     string
	Xcc                     string
	Xbcc                    string
	XFolder                 string
	XOrigin                 string
	XFileName               string
	Content                 string
}

// Email Pkg Struct
type EmailPkg struct {
	Emails []Email
}

// wait groups
var mainWg sync.WaitGroup

// emails per pkg
const emailsPerPkg = 500

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

			if len(emails) == emailsPerPkg {
				pkgChan <- EmailPkg{Emails: emails}
				emails = nil
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

// processPkgs function
func processPkgs(pkgChan <-chan EmailPkg, jsonChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		pkg, ok := <-pkgChan
		if !ok {
			// El canal de paquetes se cerró, salir del bucle
			return
		}

		// Convertir el paquete a formato JSON
		jsonData, err := json.Marshal(pkg)
		if err != nil {
			log.Printf("Error marshalling JSON: %v", err)
			continue
		}

		// Enviar el JSON al canal jsonChan
		jsonChan <- string(jsonData)
	}
}

// using marshal to convert obj into json
func toJSONString(obj interface{}) string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return ""
	}
	return string(jsonData)
}

// sendToZincSearch:
// send pkgs to zincsearch using Bulk
func sendToZincSearch(jsonData string) {
	fmt.Println("Sending package to ZincSearch...")
	req, err := http.NewRequest("POST", "http://localhost:4080/api/_bulkv2", strings.NewReader(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	client := &http.Client{
		Timeout: time.Second * 10, // Timeout de 10 segundos, ajusta según tus necesidades
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

// main function of Indexer program
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
	jsonChan := make(chan string)

	var wg sync.WaitGroup

	// go routine to process pkgs
	go func() {
		var emailsReceived int
		var emailBuffer []string

		for jsonData := range jsonChan {
			emailBuffer = append(emailBuffer, jsonData)
			emailsReceived++

			if emailsReceived == emailsPerPkg {
				// extract emails properties from email pkgs
				var records []string
				for _, emailPkgStr := range emailBuffer {
					var emailPkg EmailPkg
					if err := json.Unmarshal([]byte(emailPkgStr), &emailPkg); err == nil {
						records = append(records, toJSONString(emailPkg.Emails[0]))
					}
				}

				// parsing las json
				finalJSON := fmt.Sprintf(`{"index": "emails", "records": [%s]}`, strings.Join(records, ","))

				// sending to zincsearch
				sendToZincSearch(finalJSON)
				emailsReceived = 0
				emailBuffer = nil
			}
		}
	}()

	// Leyendo directorios y archivos
	wg.Add(1)
	go readDirectory(directoryPath, &wg, pkgChan)

	// Iniciar la goroutine para procesar paquetes
	wg.Add(1)
	go processPkgs(pkgChan, jsonChan, &wg)

	// Goroutine para cerrar el canal pkgChan después de que todas las goroutines hayan completado su trabajo
	go func() {
		wg.Wait()
		close(pkgChan)
		mainWg.Done()
	}()

	// Medir tiempo de ejecución
	startTime := time.Now()

	// Esperar a que todas las goroutines completen su trabajo
	mainWg.Wait()

	// Cerrar el canal jsonChan después de que processPkgs haya terminado
	close(jsonChan)

	// Medir tiempo de ejecución
	elapsedTime := time.Since(startTime)
	fmt.Printf("Processing took %s\n", elapsedTime)
}
