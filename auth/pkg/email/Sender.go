package email

import (
	"fmt"
	"log"
	"net/smtp"
)

func SendEmail(email string) {
  from := "me@gmail.com"
  password := "myMegaPassword"

  to := []string{
    email,
  }

  smtpHost := "smtp.gmail.com"
  smtpPort := "587"

  message := []byte("You've been hacked.")
  
  auth := smtp.PlainAuth("", from, password, smtpHost)
  
  if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message); err != nil {
    log.Fatalf("Error while trying to send email: %v", err)
  }

  fmt.Println("Email Sent Successfully!")
}

