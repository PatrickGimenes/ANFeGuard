package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strings"
)

// Config do email
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// Struct único para todos os templates
type EmailAlertData struct {
	Service  string
	CPU      string
	Memory   string
	Disk     string
	DiskPath string
	Time     string
}

func encodeSubject(subject string) string {
	return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
}

func SendEmail(cfg SMTPConfig, to []string, subject, templatePath string, data EmailAlertData) error {
	// Carregar template
	templ, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	var body bytes.Buffer

	if err := templ.Execute(&body, data); err != nil {
		return err
	}

	// Header do email
	msg := "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n"
	msg += "From: " + cfg.From + "\r\n"
	msg += "To: " + strings.Join(to, ", ") + "\r\n"
	//msg += "Subject: " + subject + "\r\n\r\n"
	msg += "Subject: " + encodeSubject(subject) + "\r\n\r\n"
	msg += body.String()

	// log.Println("FROM:", cfg.From)
	// log.Println("USER:", cfg.User)
	// log.Println("TO:", strings.Join(to, ", "))

	// Autenticação
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Envio
	if err := smtp.SendMail(addr, auth, cfg.From, to, []byte(msg)); err != nil {
		return err
	}

	log.Printf("[INFO] E-mail enviado para %v \n", to)
	return nil
}
