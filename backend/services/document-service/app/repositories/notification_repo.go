package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	NotificationID uuid.UUID
	Message        string
	Read           bool
	UserID         uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create_notification(userID uuid.UUID, message string) (*Notification, error) {
	query := `
		INSERT INTO notifications (user_id, message, read, created_at, updated_at)
		VALUES ($1, $2, false, NOW(), NOW())
		RETURNING notification_id, message, read, user_id, created_at, updated_at
	`
	n := &Notification{}
	err := r.db.QueryRow(context.Background(), query, userID, message).Scan(
		&n.NotificationID,
		&n.Message,
		&n.Read,
		&n.UserID,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	return n, err
}

func (r *NotificationRepository) Get_unread_notification(userID uuid.UUID) ([]Notification, error) {
	query := `
		SELECT notification_id, message, read, user_id, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND read = false
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(
			&n.NotificationID, &n.Message, &n.Read,
			&n.UserID, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) Get_all_notification(userID uuid.UUID) ([]Notification, error) {
	query := `
		SELECT notification_id, message, read, user_id, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(
			&n.NotificationID, &n.Message, &n.Read,
			&n.UserID, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) Mark_read_notification(notificationID uuid.UUID) error {
	_, err := r.db.Exec(
		context.Background(),
		`UPDATE notifications SET read = true, updated_at = NOW() WHERE notification_id = $1`,
		notificationID,
	)
	return err
}

func (r *NotificationRepository) Mark_all_read_notification(userID uuid.UUID) error {
	_, err := r.db.Exec(
		context.Background(),
		`UPDATE notifications SET read = true, updated_at = NOW() WHERE user_id = $1 AND read = false`,
		userID,
	)
	return err
}