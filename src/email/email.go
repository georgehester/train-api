package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type EmailClient struct {
	From        string // e.g. "you@yourdomain.com"
	Username    string // SMTP login username; if empty, From is used
	AppPassword string
	SMTPHost    string // "smtp.zoho.com"
	SMTPPort    string // "587"
}

func (client *EmailClient) Send(to []string, subject string, body string) error {
	fmt.Println(client.Username)
	fmt.Println(client.AppPassword)
	authentication := smtp.PlainAuth("", client.Username, client.AppPassword, client.SMTPHost)

	headers := map[string]string{
		"From":         client.From,
		"To":           strings.Join(to, ", "),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	var message strings.Builder
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	address := fmt.Sprintf("%s:%s", client.SMTPHost, client.SMTPPort)
	return smtp.SendMail(address, authentication, client.From, to, []byte(message.String()))
}
