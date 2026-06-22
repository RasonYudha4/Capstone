package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type EmailService struct {
	from     string
	password string
	smtpHost string
	smtpPort string
}

func NewEmailService() *EmailService {
	return &EmailService{
		from:     os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		smtpHost: os.Getenv("SMTP_HOST"),
		smtpPort: os.Getenv("SMTP_PORT"),
	}
}

func (s *EmailService) SendOTP(to string, otpCode string) error {
	subject := "Your OTP Code"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
				<h2 style="color: #6B5FAE;">Your OTP Verification Code</h2>
				<p>Please use the following 6-digit OTP code to complete your login:</p>
				<div style="font-size: 32px; font-weight: bold; letter-spacing: 5px; color: #6B5FAE; margin: 20px 0; padding: 15px; background-color: #f1eeff; border-radius: 5px; text-align: center; display: inline-block; min-width: 200px;">
					%s
				</div>
				<p style="color: #666; font-size: 14px;">This code is valid for 5 minutes. Do not share this code with anyone.</p>
			</div>
		</body>
		</html>
	`, otpCode)

	log.Printf("[email] SendOTP: attempting to send to %s via %s:%s", to, s.smtpHost, s.smtpPort)

	auth := smtp.PlainAuth("", s.from, s.password, s.smtpHost)

	conn, err := smtp.Dial(s.smtpHost + ":" + s.smtpPort)
	if err != nil {
		log.Printf("[email] Dial FAILED: %v", err)
		return err
	}
	defer conn.Quit()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.smtpHost,
	}

	if err = conn.StartTLS(tlsConfig); err != nil {
		log.Printf("[email] StartTLS FAILED: %v", err)
		return err
	}
	log.Printf("[email] StartTLS SUCCESS")

	if err = conn.Auth(auth); err != nil {
		log.Printf("[email] Auth FAILED: %v", err)
		return err
	}
	log.Printf("[email] Auth SUCCESS")

	if err = conn.Mail(s.from); err != nil {
		log.Printf("[email] Mail from FAILED: %v", err)
		return err
	}

	if err = conn.Rcpt(to); err != nil {
		log.Printf("[email] Rcpt to %s FAILED: %v", to, err)
		return err
	}

	w, err := conn.Data()
	if err != nil {
		log.Printf("[email] Data FAILED: %v", err)
		return err
	}

	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		body

	if _, err = w.Write([]byte(msg)); err != nil {
		log.Printf("[email] Write FAILED: %v", err)
		return err
	}

	if err = w.Close(); err != nil {
		log.Printf("[email] Close FAILED: %v", err)
		return err
	}

	log.Printf("[email] SendOTP SUCCESS to %s", to)
	return nil
}