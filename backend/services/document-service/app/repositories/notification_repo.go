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
	// Hack for live DB: Add column if missing
	_, _ = db.Exec(context.Background(), "ALTER TABLE notifications ADD COLUMN IF NOT EXISTS document_id UUID REFERENCES documents(document_id) ON DELETE CASCADE")
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create_notification(userID uuid.UUID, message string, documentID uuid.UUID) (*Notification, error) {
	query := `
		INSERT INTO notifications (user_id, message, document_id, read, created_at, updated_at)
		VALUES ($1, $2, CASE WHEN $3 = '00000000-0000-0000-0000-000000000000'::uuid THEN NULL ELSE $3 END, false, NOW(), NOW())
		RETURNING notification_id, message, read, user_id, created_at, updated_at
	`
	n := &Notification{}
	err := r.db.QueryRow(context.Background(), query, userID, message, documentID).Scan(
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
		SELECT n.notification_id, 
		       CASE 
		           WHEN n.message = 'new_document' AND n.document_id IS NOT NULL THEN 
		               'A new document (' || d.filename || ' by ' || u.email || ') requires your approval'
		           WHEN n.message = 'document_status_update' AND n.document_id IS NOT NULL THEN
		               'Your Document (' || d.filename || ' by ' || u.email || ') is now ' || d.status
		           ELSE n.message 
		       END as message,
		       n.read, n.user_id, n.created_at, n.updated_at
		FROM notifications n
		LEFT JOIN documents d ON n.document_id = d.document_id
		LEFT JOIN users u ON d.created_by = u.user_id
		WHERE n.user_id = $1 AND n.read = false
		ORDER BY n.created_at DESC
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
		SELECT n.notification_id, 
		       CASE 
		           WHEN n.message = 'new_document' AND n.document_id IS NOT NULL THEN 
		               'A new document (' || d.filename || ' by ' || u.email || ') requires your approval'
		           WHEN n.message = 'document_status_update' AND n.document_id IS NOT NULL THEN
		               'Your Document (' || d.filename || ' by ' || u.email || ') is now ' || d.status
		           ELSE n.message 
		       END as message,
		       n.read, n.user_id, n.created_at, n.updated_at
		FROM notifications n
		LEFT JOIN documents d ON n.document_id = d.document_id
		LEFT JOIN users u ON d.created_by = u.user_id
		WHERE n.user_id = $1
		ORDER BY n.created_at DESC
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