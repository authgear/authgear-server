package main

import (
	"bytes"
	"testing"

	"gopkg.in/gomail.v2"
)

func TestParseMail(t *testing.T) {
	// Built the same way as pkg/lib/infra/mail.Sender.PrepareMessage.
	msg := gomail.NewMessage()
	msg.SetHeader("From", "no-reply@example.com")
	msg.SetHeader("To", "user@example.com")
	msg.SetHeader("Subject", "登入 Authgear")
	link := "https://custom.example.com/login-link?a=b&code=M6CGA4WV6M9XTXNWFYFHRQDWF6VFR7K4"
	msg.SetBody("text/plain", "Open this link to sign in:\n"+link+"\n")
	msg.AddAlternative("text/html", `<a href="`+link+`">Sign in</a>`)

	var buf bytes.Buffer
	if _, err := msg.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	entry, err := parseMail(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if entry.Subject != "登入 Authgear" {
		t.Errorf("subject: got %q", entry.Subject)
	}
	if len(entry.To) != 1 || entry.To[0] != "user@example.com" {
		t.Errorf("to: got %v", entry.To)
	}
	if !bytes.Contains([]byte(entry.Text), []byte(link)) {
		t.Errorf("text does not contain the link: %q", entry.Text)
	}
}
