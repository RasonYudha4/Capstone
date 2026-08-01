package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"time"

	"auth-service/config"
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
	invitationLink := fmt.Sprintf("%s/setup-password?token=%s", config.FrontendURL, token)
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

func (s *EmailService) SendPasswordReset(to string, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", config.FrontendURL, token)
	subject := "Reset Password"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
				<h2 style="color: #6B5FAE;">Reset Your Password</h2>
				<p>We received a request to reset your password. Click the button below to choose a new password:</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="%s" style="background-color: #6B5FAE; color: #ffffff; padding: 12px 25px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block;">
						Reset Password
					</a>
				</div>
				<p style="color: #666; font-size: 14px;">This link is valid for 1 hour and can only be used once. If you did not request a password reset, please ignore this email.</p>
				<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
				<p style="font-size: 12px; color: #999;">If the button doesn't work, copy and paste this link into your browser:<br>
				<a href="%s" style="color: #6B5FAE;">%s</a></p>
			</div>
		</body>
		</html>
	`, resetLink, resetLink, resetLink)

	log.Printf("[email] SendPasswordReset: attempting to send to %s", to)
	return s.sendRawEmail(to, subject, body)
}

func (s *EmailService) SendPasswordChangedNotice(to string) error {
	subject := "Password Changed"
	body := `
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
				<h2 style="color: #6B5FAE;">Password Updated</h2>
				<p>Your password was changed successfully. All active sessions have been signed out.</p>
				<p style="color: #666; font-size: 14px;">If you did not make this change, contact your system administrator immediately.</p>
			</div>
		</body>
		</html>
	`

	log.Printf("[email] SendPasswordChangedNotice: attempting to send to %s", to)
	return s.sendRawEmail(to, subject, body)
}

func (s *EmailService) sendRawEmail(to string, subject string, body string) error {
	if s.from == "" || s.password == "" || s.smtpHost == "" || s.smtpPort == "" {
		log.Printf("[email] Skipping email send: SMTP credentials not fully configured")
		if config.IsDevMode {
			return nil
		}
		return fmt.Errorf("SMTP is not configured")
	}

	auth := smtp.PlainAuth("", s.from, s.password, s.smtpHost)

	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", s.smtpHost+":"+s.smtpPort)
	if err != nil {
		log.Printf("[email] Dial FAILED: %v", err)
		return err
	}

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		conn.Close()
		log.Printf("[email] NewClient FAILED: %v", err)
		return err
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.smtpHost,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("[email] StartTLS FAILED: %v", err)
		return err
	}

	if err = client.Auth(auth); err != nil {
		log.Printf("[email] Auth FAILED: %v", err)
		return err
	}

	if err = client.Mail(s.from); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
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