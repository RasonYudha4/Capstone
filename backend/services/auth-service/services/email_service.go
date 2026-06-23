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
		from:     os.Getenv("SMTP_FROM"),     // Google requires this to match auth user
		password: os.Getenv("SMTP_PASSWORD"), // App Password
		smtpHost: os.Getenv("SMTP_HOST"),     // smtp.gmail.com
		smtpPort: os.Getenv("SMTP_PORT"),     // 587
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

	log.Printf("[email] SendOTP: attempting to send to %s", to)
	return s.sendRawEmail(to, subject, body)
}

func (s *EmailService) SendInvitation(to string, role string, token string) error {
	invitationLink := fmt.Sprintf("http://localhost:5173/setup-password?token=%s", token)
	subject := "Invitation to Join Capstone System"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
				<h2 style="color: #6B5FAE;">Welcome to Capstone!</h2>
				<p>You have been invited to join the system as a <strong>%s</strong>.</p>
				<p>Please click the button below to set up your account and password:</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="%s" style="background-color: #6B5FAE; color: #ffffff; padding: 12px 25px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block;">
						Set Up Account
					</a>
				</div>
				<p style="color: #666; font-size: 14px;">This link is valid for 24 hours. If you did not expect this invitation, please ignore this email.</p>
				<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
				<p style="font-size: 12px; color: #999;">If the button doesn't work, copy and paste this link into your browser:<br>
				<a href="%s" style="color: #6B5FAE;">%s</a></p>
			</div>
		</body>
		</html>
	`, role, invitationLink, invitationLink, invitationLink)

	log.Printf("[email] SendInvitation: attempting to send to %s", to)
	return s.sendRawEmail(to, subject, body)
}

func (s *EmailService) sendRawEmail(to string, subject string, body string) error {
	if s.from == "" || s.password == "" || s.smtpHost == "" || s.smtpPort == "" {
		log.Printf("[email] Skiping email send: SMTP credentials not fully configured")
		return nil // In dev mode, we might not have SMTP configured
	}

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

	if err = conn.Auth(auth); err != nil {
		log.Printf("[email] Auth FAILED: %v", err)
		return err
	}

	if err = conn.Mail(s.from); err != nil {
		return err
	}

	if err = conn.Rcpt(to); err != nil {
		return err
	}

	w, err := conn.Data()
	if err != nil {
		return err
	}

	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		body

	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}

	log.Printf("[email] Email sent successfully to %s", to)
	return w.Close()
}