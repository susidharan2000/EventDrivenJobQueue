package main

import (
	"encoding/json"
	"os"

	"gopkg.in/gomail.v2"
)

func sendMail(payload json.RawMessage) error {
	var email Email
	err := json.Unmarshal(payload, &email)
	if err != nil {
		panic(err)
	}

	user := os.Getenv("GMAIL_USER")
	pass := os.Getenv("GMAIL_APP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if user == "" || pass == "" || host == "" || port == "" {
		panic("Missing email env variables")
	}

	mail := gomail.NewMessage()
	mail.SetHeader("From", user)
	mail.SetHeader("To", email.Email)
	mail.SetHeader("Subject", email.Subject)
	mail.SetBody("text/plain", email.Body)

	dial := gomail.NewDialer(
		host,
		587,
		user,
		pass,
	)
	if err := dial.DialAndSend(mail); err != nil {
		return err
	}
	//fmt.Println("Mail Sent Successfully")
	return nil
}
