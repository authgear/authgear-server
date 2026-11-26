package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"strings"
	"sync"

	"github.com/phires/go-guerrilla/backends"
	guerrillamail "github.com/phires/go-guerrilla/mail"
)

// MailLogPath is read by the e2e test runner. It is relative to the e2e
// directory, where run.sh starts this server.
const MailLogPath = "./logs/e2e-smtp-mail.jsonl"

// MailLogEntry is one line of MailLogPath.
type MailLogEntry struct {
	Subject string   `json:"subject"`
	To      []string `json:"to"`
	Text    string   `json:"text"`
}

// MailLogProcessor appends the decoded text/plain body of every mail to
// MailLogPath, so that a test can assert on the content, such as a link.
func MailLogProcessor() backends.Decorator {
	var mu sync.Mutex
	return func(p backends.Processor) backends.Processor {
		return backends.ProcessWith(func(e *guerrillamail.Envelope, task backends.SelectTask) (backends.Result, error) {
			if task == backends.TaskSaveMail {
				entry, err := parseMail(e.Data.Bytes())
				if err != nil {
					backends.Log().WithError(err).Error("failed to parse mail for the mail log")
				} else {
					mu.Lock()
					err = appendMailLog(entry)
					mu.Unlock()
					if err != nil {
						backends.Log().WithError(err).Error("failed to write the mail log")
					}
				}
			}
			return p.Process(e, task)
		})
	}
}

func appendMailLog(entry *MailLogEntry) error {
	f, err := os.OpenFile(MailLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

func parseMail(data []byte) (*MailLogEntry, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	decoder := new(mime.WordDecoder)
	subject, err := decoder.DecodeHeader(msg.Header.Get("Subject"))
	if err != nil {
		return nil, err
	}

	var to []string
	if addrs, err := msg.Header.AddressList("To"); err == nil {
		for _, addr := range addrs {
			to = append(to, addr.Address)
		}
	}

	text, err := findText(msg.Header.Get("Content-Type"), msg.Header.Get("Content-Transfer-Encoding"), msg.Body)
	if err != nil {
		return nil, err
	}

	return &MailLogEntry{Subject: subject, To: to, Text: text}, nil
}

// findText returns the first text/plain part, decoded.
func findText(contentType string, transferEncoding string, body io.Reader) (string, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		r := multipart.NewReader(body, params["boundary"])
		for {
			part, err := r.NextRawPart()
			if err == io.EOF {
				return "", nil
			}
			if err != nil {
				return "", err
			}
			text, err := findText(part.Header.Get("Content-Type"), part.Header.Get("Content-Transfer-Encoding"), part)
			if err != nil {
				return "", err
			}
			if text != "" {
				return text, nil
			}
		}
	}

	if mediaType != "text/plain" {
		return "", nil
	}

	var decoded io.Reader
	switch strings.ToLower(transferEncoding) {
	case "quoted-printable":
		decoded = quotedprintable.NewReader(body)
	case "base64":
		decoded = base64.NewDecoder(base64.StdEncoding, body)
	case "", "7bit", "8bit", "binary":
		decoded = body
	default:
		return "", fmt.Errorf("unsupported content-transfer-encoding: %s", transferEncoding)
	}

	b, err := io.ReadAll(decoded)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
