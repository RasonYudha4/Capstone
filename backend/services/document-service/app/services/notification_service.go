package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"sync"

	"capstone/app/repositories"

	"github.com/google/uuid"
)

type SSEEvent struct {
	Type           string `json:"type"`
	NotificationId string `json:"notification_id,omitempty"`
	DocumentId     string `json:"document_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

type connKey struct {
	userID string
	connID string
}

type NotificationService struct {
	from          string
	password      string
	smtpHost      string
	smtpPort      string
	notifications *repositories.NotificationRepository
	mu            sync.RWMutex
	clients       map[connKey]chan SSEEvent
}

func NewNotificationService(notifRepo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{
		from:          os.Getenv("SMTP_USERNAME"),
		password:      os.Getenv("SMTP_PASSWORD"),
		smtpHost:      os.Getenv("SMTP_HOST"),
		smtpPort:      os.Getenv("SMTP_PORT"),
		notifications: notifRepo,
		clients:       make(map[connKey]chan SSEEvent),
	}
}

func (n *NotificationService) RegisterSSE(userID, connID string, ch chan SSEEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.clients[connKey{userID, connID}] = ch
	log.Printf("[SSE] registered   user=%s conn=%s (user tabs=%d total=%d)",
		userID, connID, n.countUserLocked(userID), len(n.clients))
}

func (n *NotificationService) UnregisterSSE(userID, connID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	k := connKey{userID, connID}
	if ch, ok := n.clients[k]; ok {
		close(ch)
		delete(n.clients, k)
	}
	log.Printf("[SSE] unregistered user=%s conn=%s (user tabs=%d total=%d)",
		userID, connID, n.countUserLocked(userID), len(n.clients))
}

func (n *NotificationService) countUserLocked(userID string) int {
	c := 0
	for k := range n.clients {
		if k.userID == userID {
			c++
		}
	}
	return c
}

func (n *NotificationService) NotifySSE(userIds []uuid.UUID, event SSEEvent) {
	for _, uid := range userIds {
		saved, err := n.notifications.Create_notification(uid, event.Message)
		if err != nil {
			log.Printf("[SSE] NotifySSE: DB save failed for %s: %v", uid, err)
			continue
		}
		log.Printf("[SSE] NotifySSE: saved notification %s for user %s", saved.NotificationID, uid)
	}

	n.mu.RLock()
	var targets []chan SSEEvent
	for k, ch := range n.clients {
		for _, uid := range userIds {
			if k.userID == uid.String() {
				targets = append(targets, ch)
				break
			}
		}
	}
	n.mu.RUnlock()

	if len(targets) == 0 {
		log.Printf("[SSE] NotifySSE: all users offline — saved to DB only")
		return
	}

	for _, ch := range targets {
		select {
		case ch <- event:
		default:
			log.Printf("[SSE] NotifySSE: channel full, dropping %q", event.Type)
		}
	}
	log.Printf("[SSE] NotifySSE: delivered %q to %d tab(s)", event.Type, len(targets))
}

func (n *NotificationService) SendEmail(to []string, subject, body string) error {
	log.Printf("[email] SendEmail: attempting to send to %v via %s:%s", to, n.smtpHost, n.smtpPort)

	auth := smtp.PlainAuth("", n.from, n.password, n.smtpHost)

	conn, err := smtp.Dial(n.smtpHost + ":" + n.smtpPort)
	if err != nil {
		log.Printf("[email] Dial FAILED: %v", err)
		return err
	}
	defer conn.Quit()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         n.smtpHost,
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

	if err = conn.Mail(n.from); err != nil {
		log.Printf("[email] Mail from FAILED: %v", err)
		return err
	}

	for _, addr := range to {
		if err = conn.Rcpt(addr); err != nil {
			log.Printf("[email] Rcpt to %s FAILED: %v", addr, err)
			return err
		}
	}

	w, err := conn.Data()
	if err != nil {
		log.Printf("[email] Data FAILED: %v", err)
		return err
	}

	msg := "From: " + n.from + "\r\n" +
		"To: " + strings.Join(to, ", ") + "\r\n" +
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

	log.Printf("[email] SendEmail SUCCESS to %v", to)
	return nil
}

func (n *NotificationService) NotifyOwner(email string, documentName string, status string) {
		subject := "Document Status Updated"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
				<h2 style="color: #6B5FAE;">Document Status Updated</h2>
				<p>Your document status has been updated:</p>
				<div style="margin: 20px 0; padding: 15px; background-color: #f1eeff; border-radius: 5px; text-align: center;">
					<p style="font-size: 18px; font-weight: bold; color: #6B5FAE; margin: 0 0 10px 0;">%s</p>
					<p style="font-size: 16px; color: #444; margin: 0;">Status: <span style="font-weight: bold; color: %s;">%s</span></p>
				</div>
				<p style="color: #666; font-size: 14px;">Please log in to the system to view your document details.</p>
			</div>
		</body>
		</html>
	`, documentName, statusColor(status), status)
	if err := n.SendEmail([]string{email}, subject, body); err != nil {
		log.Printf("[email] NotifyDeptHead failed for %s: %v", email, err)
	}
}

func (n *NotificationService) NotifyDeptHead(email string, documentName string) {
	subject := "New Document For Approval"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
				<h2 style="color: #6B5FAE;">New Document Requires Approval</h2>
				<p>A new document has been submitted and requires your review:</p>
				<div style="font-size: 20px; font-weight: bold; color: #6B5FAE; margin: 20px 0; padding: 15px; background-color: #f1eeff; border-radius: 5px; text-align: center;">
					%s
				</div>
				<p style="color: #666; font-size: 14px;">Please log in to the system to review and approve or reject this document.</p>
			</div>
		</body>
		</html>
	`, documentName)
	if err := n.SendEmail([]string{email}, subject, body); err != nil {
		log.Printf("[email] NotifyDeptHead failed for %s: %v", email, err)
	}
}

func statusColor(status string) string {
	switch status {
	case "Approved":
		return "#22c55e"
	case "Rejected":
		return "#ef4444" 
	default:
		return "#f59e0b" 
	}
}