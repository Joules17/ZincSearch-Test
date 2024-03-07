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
			// Extra procedure for Message Id
			if prefix == "Message-ID: " {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					messageID := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(parts[1], ".JavaMail.evans@thyme"), ".JavaMail.evans@thyme"))
					if messageID == "" {
						// Null file
						log.Println("Invalid MessageID in file ", fileName)
					}
					*field = messageID
				}
			} else {
				// Same procedure for other fields
				*field = strings.TrimSpace(strings.TrimPrefix(line, prefix))
			}
			return
		}
	}

	// If no prefix matches, it's email content text
	email.Content += line + "\n"
}
