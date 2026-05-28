package services

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"sync"
	"crypto/tls"

	"github.com/google/uuid"
	"capstone/app/repositories"
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

	mu      sync.RWMutex
	clients map[connKey]chan SSEEvent
}

func NewNotificationService(notifRepo *repositories.NotificationRepository, apiKey string, fromAddress string) *NotificationService {
	return &NotificationService{
		from:          fromAddress,
		password:      apiKey,
		smtpHost:      "smtp.resend.com",
		smtpPort:      "465",
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

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         n.smtpHost,
	}

	conn, err := tls.Dial("tcp", n.smtpHost+":"+n.smtpPort, tlsConfig)
	if err != nil {
		log.Printf("[email] TLS dial FAILED: %v", err)
		return err
	}
	log.Printf("[email] TLS connection SUCCESS")

	c, err := smtp.NewClient(conn, n.smtpHost)
	if err != nil {
		log.Printf("[email] SMTP client FAILED: %v", err)
		return err
	}
	defer c.Quit()

	auth := smtp.PlainAuth("", "resend", n.password, n.smtpHost)
	if err = c.Auth(auth); err != nil {
		log.Printf("[email] Auth FAILED: %v", err)
		return err
	}
	log.Printf("[email] Auth SUCCESS")

	if err = c.Mail(n.from); err != nil {
		log.Printf("[email] Mail from FAILED: %v", err)
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			log.Printf("[email] Rcpt to %s FAILED: %v", addr, err)
			return err
		}
	}

	w, err := c.Data()
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

func (n *NotificationService) NotifyOwner(email string, documentId uuid.UUID, status string) {
	subject := "Document Status Updated"
	body := fmt.Sprintf("Your document (%s) status has been updated to: %s", documentId, status)
	if err := n.SendEmail([]string{email}, subject, body); err != nil {
		log.Printf("[email] NotifyOwner failed for %s: %v", email, err)
	}
}


func (n *NotificationService) NotifyDeptHead(emails []string, documentName string) {
	subject := "New Document For Approval"
	body := fmt.Sprintf("A new document requires your approval.\n\nDocument: %s", documentName)
	if err := n.SendEmail(emails, subject, body); err != nil {
		log.Printf("[email] NotifyDeptHead failed for %s: %v", strings.Join(emails, ", "), err)
	}
}