package repositories

import (
	"capstone/app/schemas"
	"context"
	"database/sql"
    "log"
	"github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5"
)

type FormOptionsRepository interface {
	GetAllNested(ctx context.Context, userID string) ([]schemas.ServiceOption, error)
	GetDocumentTypes(ctx context.Context) ([]schemas.DocumentTypeOption, error)
}

type formOptionsRepo struct {
	db *pgxpool.Pool
}

func NewFormOptionsRepository(db *pgxpool.Pool) FormOptionsRepository {
	return &formOptionsRepo{db: db}
}

func (f *formOptionsRepo) GetAllNested(ctx context.Context, userID string) ([]schemas.ServiceOption, error) {
	var role, groupID string
	err := f.db.QueryRow(ctx, `
		SELECT role, group_id FROM users WHERE user_id = $1
	`, userID).Scan(&role, &groupID)
	if err != nil {
		return nil, err
	}

	baseQuery := `
		SELECT
			sv.service_id, sv.service_code, sv.description,
			st.standard_id, st.standard_code, st.description,
			a.assessment_id, a.assessment_code, a.description
		FROM services sv
		LEFT JOIN standard   st ON st.service_id = sv.service_id
		LEFT JOIN assessment a  ON a.standard_id = st.standard_id
	`
    log.Print(userID)

	var rows pgx.Rows
	if role == "master-admin" {
		rows, err = f.db.Query(ctx, baseQuery)
	} else {
		rows, err = f.db.Query(ctx, baseQuery+` WHERE sv.group_id = $1`, groupID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	serviceMap := make(map[string]*schemas.ServiceOption)
	serviceOrder := []string{}

	for rows.Next() {
		var (
			svcID, svcCode, svcDesc string
			stdID, stdCode, stdDesc sql.NullString
			asmID, asmCode, asmDesc sql.NullString
		)

		if err := rows.Scan(
			&svcID, &svcCode, &svcDesc,
			&stdID, &stdCode, &stdDesc,
			&asmID, &asmCode, &asmDesc,
		); err != nil {
			return nil, err
		}

		if _, exists := serviceMap[svcID]; !exists {
			serviceMap[svcID] = &schemas.ServiceOption{
				ID: svcID, Code: svcCode, Description: svcDesc,
				Standards: []schemas.StandardOption{},
			}
			serviceOrder = append(serviceOrder, svcID)
		}

		if !stdID.Valid {
			continue
		}

		svc := serviceMap[svcID]
		stdIndex := -1
		for i, s := range svc.Standards {
			if s.ID == stdID.String {
				stdIndex = i
				break
			}
		}
		if stdIndex == -1 {
			svc.Standards = append(svc.Standards, schemas.StandardOption{
				ID: stdID.String, Code: stdCode.String, Description: stdDesc.String,
				Assessments: []schemas.AssessmentOption{},
			})
			stdIndex = len(svc.Standards) - 1
		}

		if !asmID.Valid {
			continue
		}

		svc.Standards[stdIndex].Assessments = append(
			svc.Standards[stdIndex].Assessments,
			schemas.AssessmentOption{ID: asmID.String, Code: asmCode.String, Description: asmDesc.String},
		)
	}

	result := make([]schemas.ServiceOption, 0, len(serviceOrder))
	for _, id := range serviceOrder {
		result = append(result, *serviceMap[id])
	}
	return result, nil
}

func (f *formOptionsRepo) GetDocumentTypes(ctx context.Context) ([]schemas.DocumentTypeOption, error) {
	rows, err := f.db.Query(ctx, `
		SELECT document_type_id, name, description
		FROM document_types
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []schemas.DocumentTypeOption
	for rows.Next() {
		var dt schemas.DocumentTypeOption
		if err := rows.Scan(&dt.ID, &dt.Name, &dt.Description); err != nil {
			return nil, err
		}
		result = append(result, dt)
	}
	return result, nil
}