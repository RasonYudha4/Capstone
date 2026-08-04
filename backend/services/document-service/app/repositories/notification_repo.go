package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	NotificationID     uuid.UUID  `json:"NotificationID"`
	Message            string     `json:"Message"`
	Read               bool       `json:"Read"`
	UserID             uuid.UUID  `json:"UserID"`
	CreatedAt          time.Time  `json:"CreatedAt"`
	UpdatedAt          time.Time  `json:"UpdatedAt"`
	DocumentID         *uuid.UUID `json:"DocumentID,omitempty"`
	Filename           string     `json:"Filename,omitempty"`
	DocumentType       string     `json:"DocumentType,omitempty"`
	CreatedBy          string     `json:"CreatedBy,omitempty"`
	DocumentStatus     string     `json:"DocumentStatus,omitempty"`
	DocumentUpdatedAt  *time.Time `json:"DocumentUpdatedAt,omitempty"`
	ServiceCode        string     `json:"ServiceCode,omitempty"`
	StandardCode       string     `json:"StandardCode,omitempty"`
	AssessmentCode     string     `json:"AssessmentCode,omitempty"`
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
	if err == nil && documentID != uuid.Nil {
		id := documentID
		n.DocumentID = &id
	}
	return n, err
}

const notificationSelect = `
	SELECT n.notification_id,
	       CASE
	           WHEN n.message = 'new_document' AND n.document_id IS NOT NULL THEN
	               'A new document (' || COALESCE(d.filename, 'unknown') || ' by ' || COALESCE(u.email, 'unknown') || ') requires your approval'
	           WHEN n.message = 'document_status_update' AND n.document_id IS NOT NULL THEN
	               'Your Document (' || COALESCE(d.filename, 'unknown') || ' by ' || COALESCE(u.email, 'unknown') || ') is now ' || COALESCE(d.status, 'unknown')
	           ELSE n.message
	       END as message,
	       n.read, n.user_id, n.created_at, n.updated_at,
	       n.document_id,
	       COALESCE(d.filename, ''),
	       COALESCE(dt.name, ''),
	       COALESCE(u.email, ''),
	       COALESCE(d.status, ''),
	       d.updated_at,
	       COALESCE(s.service_code, ''),
	       COALESCE(st.standard_code, ''),
	       COALESCE(a.assessment_code, '')
	FROM notifications n
	LEFT JOIN documents d ON n.document_id = d.document_id
	LEFT JOIN document_types dt ON d.document_type_id = dt.document_type_id
	LEFT JOIN users u ON d.created_by = u.user_id
	LEFT JOIN services s ON d.service_id = s.service_id
	LEFT JOIN standard st ON d.standard_id = st.standard_id
	LEFT JOIN assessment a ON d.assessment_id = a.assessment_id
`

func scanNotifications(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]Notification, error) {
	var notifications []Notification
	for rows.Next() {
		var n Notification
		var docUpdatedAt *time.Time
		if err := rows.Scan(
			&n.NotificationID, &n.Message, &n.Read,
			&n.UserID, &n.CreatedAt, &n.UpdatedAt,
			&n.DocumentID,
			&n.Filename,
			&n.DocumentType,
			&n.CreatedBy,
			&n.DocumentStatus,
			&docUpdatedAt,
			&n.ServiceCode,
			&n.StandardCode,
			&n.AssessmentCode,
		); err != nil {
			return nil, err
		}
		n.DocumentUpdatedAt = docUpdatedAt
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) Get_unread_notification(userID uuid.UUID) ([]Notification, error) {
	query := notificationSelect + `
		WHERE n.user_id = $1 AND n.read = false
		ORDER BY n.created_at DESC
	`
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotifications(rows)
}

func (r *NotificationRepository) Get_all_notification(userID uuid.UUID) ([]Notification, error) {
	query := notificationSelect + `
		WHERE n.user_id = $1
		ORDER BY n.created_at DESC
	`
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotifications(rows)
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
