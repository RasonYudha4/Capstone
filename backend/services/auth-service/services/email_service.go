package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

type EmailService struct {
	from     string
	password string
	smtpHost string
	smtpPort string
}

func NewEmailService(from, password string) *EmailService {
	return &EmailService{
		from:     from,
		password: password,
		smtpHost: "smtp.resend.com",
		smtpPort: "465",
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

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.smtpHost,
	}

	conn, err := tls.Dial("tcp", s.smtpHost+":"+s.smtpPort, tlsConfig)
	if err != nil {
		log.Printf("[email] TLS dial FAILED: %v", err)
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		log.Printf("[email] SMTP client FAILED: %v", err)
		return err
	}
	defer c.Quit()

	auth := smtp.PlainAuth("", "resend", s.password, s.smtpHost)
	if err = c.Auth(auth); err != nil {
		log.Printf("[email] Auth FAILED: %v", err)
		return err
	}

	if err = c.Mail(s.from); err != nil {
		log.Printf("[email] Mail from FAILED: %v", err)
		return err
	}

	if err = c.Rcpt(to); err != nil {
		log.Printf("[email] Rcpt to %s FAILED: %v", to, err)
		return err
	}

	w, err := c.Data()
	if err != nil {
		log.Printf("[email] Data FAILED: %v", err)
		return err
	}
	defer w.Close()

	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		body

	if _, err = w.Write([]byte(msg)); err != nil {
		log.Printf("[email] Write FAILED: %v", err)
		return err
	}

	log.Printf("[email] SendOTP SUCCESS to %s", to)
	return nil
}
