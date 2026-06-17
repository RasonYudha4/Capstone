package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"capstone/app/repositories"
	"capstone/app/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
	notificationRepo    *repositories.NotificationRepository
}

func NewNotificationHandler(svc *services.NotificationService, repo *repositories.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{notificationService: svc, notificationRepo: repo}
}

func (h *NotificationHandler) SSEHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	connID := uuid.NewString()

	log.Printf("[SSE] client connected: user=%s conn=%s", userID, connID)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	ch := make(chan services.SSEEvent, 10)
	h.notificationService.RegisterSSE(userID, connID, ch)
	defer func() {
		h.notificationService.UnregisterSSE(userID, connID)
		log.Printf("[SSE] client disconnected: user=%s conn=%s", userID, connID)
	}()

	uid, _ := uuid.Parse(userID)
	missed, err := h.notificationRepo.Get_unread_notification(uid)
	if err != nil {
		log.Printf("[SSE] failed to fetch unread for user=%s: %v", userID, err)
	}
	if len(missed) > 0 {
		type syncPayload struct {
			Type  string `json:"type"`
			Count int    `json:"count"`
		}
		data, _ := json.Marshal(syncPayload{Type: "sync", Count: len(missed)})
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()
	}

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			flusher.Flush()
			log.Printf("[SSE] sent %q to user=%s conn=%s", event.Type, userID, connID)

		case <-time.After(30 * time.Second):
			fmt.Fprintf(c.Writer, ": ping\n\n")
			flusher.Flush()

		case <-c.Request.Context().Done():
			return
		}
	}
}

func (h *NotificationHandler) GetAll(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	notifications, err := h.notificationRepo.Get_all_notification(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": notifications})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}
	if err := h.notificationRepo.Mark_read_notification(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	if err := h.notificationRepo.Mark_all_read_notification(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all marked as read"})
}