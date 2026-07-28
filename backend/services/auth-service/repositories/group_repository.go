package repositories

import (
	"auth-service/models"
	"database/sql"
	"fmt"
)

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) GetAll() ([]models.Group, error) {
	rows, err := r.db.Query(
		`SELECT DISTINCT ON (LOWER(group_name)) group_id, group_name, created_at, updated_at
		 FROM groups
		 ORDER BY LOWER(group_name), created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query groups: %w", err)
	}
	defer rows.Close()

	groups := make([]models.Group, 0)
	for rows.Next() {
		var g models.Group
		if err := rows.Scan(&g.GroupID, &g.GroupName, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan group row: %w", err)
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group rows: %w", err)
	}
	return groups, nil
}

func (r *GroupRepository) Exists(groupID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM groups WHERE group_id = $1)`, groupID,
	).Scan(&exists)
	return exists, err
}
