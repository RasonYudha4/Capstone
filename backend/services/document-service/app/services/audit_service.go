package services

import 
(
	"capstone/app/schemas"
	"capstone/app/repositories"
)

type AuditService struct {
	repo *repositories.AuditRepo
}

func NewAuditService(repo *repositories.AuditRepo) *AuditService{
	return &AuditService{
		repo: repo,
	}
}

func (a *AuditService) Get_audit()([]schemas.AuditResponse, error){
	audit, err := a.repo.GetAudit()
	if err != nil{
		return []schemas.AuditResponse{}, err
	}

	return audit, nil

}