package Indexer

import (
	"bufio"
	"log"
	"strings"
)

// scan file function
// scan file content and return an EmailObject
func ScanFile(fileName, fileContent string) (Email, error) {
	// email struct
	email := Email{}

	// scanner
	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	for scanner.Scan() {
		line := scanner.Text()
		ParseLine(line, &email, fileName)
	}

	// fmt.Println("Scanned file successfully: ", fileName)
	return email, nil
}

// parseline function
// assigns the value to email structure field
func ParseLine(line string, email *Email, fileName string) {
	// prefix map
	prefixMap := map[string]*string{
		"Message-ID: ":                &email.MessageID,
		"Date: ":                      &email.Date,
		"From: ":                      &email.From,
		"To: ":                        &email.To,
		"Subject: ":                   &email.Subject,
		"Mime-Version: ":              &email.MimeVersion,
		"Content-Type: ":              &email.ContentType,
		"Content-Transfer-Encoding: ": &email.ContentTransferEncoding,
		"X-From: ":                    &email.XFrom,
		"X-To: ":                      &email.XTo,
		"X-cc: ":                      &email.Xcc,
		"X-bcc: ":                     &email.Xbcc,
		"X-Folder: ":                  &email.XFolder,
		"X-Origin: ":                  &email.XOrigin,
		"X-FileName: ":                &email.XFileName,
	}

	for prefix, field := range prefixMap {
		if strings.HasPrefix(line, prefix) {
			// extra procedure for Message Id
			if prefix == "Message-ID: " {
				idStart := strings.Index(line, "<") + 1
				idEnd := strings.Index(line, ">")
				if idStart != -1 && idEnd != -1 && idEnd > idStart {
					messageID := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line[idStart:idEnd], ".JavaMail.evans@thyme"), ".JavaMail.evans@thyme"))
					if messageID == "" {
						// Null file
						log.Println("Invalid MessageID in file ", fileName)
					}
					*field = messageID
				}
			} else {
				// same procedure for other fields
				*field = strings.TrimSpace(strings.TrimPrefix(line, prefix))
			}
			return
		}
	}

	// dont match with any prefix? Then it's email content text
	email.Content += line + "\n"
}
