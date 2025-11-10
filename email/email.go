package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
)

// Config email
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// Dados do e-mail
type EmailData struct {
	CPU    float64
	Memory float64
	Disk   float64
	Time   string
}

func SendEmail(cfg SMTPConfig, to []string, subject, templatePath string, data EmailData) error {
	//carrega o template
	templ, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	var body bytes.Buffer

	//render do template
	if err := templ.Execute(&body, data); err != nil {
		return err
	}

	//Monta o header do e-mail
	msg := "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n"
	msg += "From: " + cfg.From + "\r\n"
	msg += "To: " + to[0] + "\r\n"
	msg += "Subject: " + subject + "\r\n\r\n"
	msg += body.String()

	//login e envio
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Password)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	if err := smtp.SendMail(addr, auth, cfg.From, to, []byte(msg)); err != nil {
		return err
	}
	log.Printf("E-mail enviado para %v \n", to)

	return nil
}
