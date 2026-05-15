package services

import (
    "capstone/app/repositories"
    "capstone/app/schemas"
    "context"
)

type FormOptionsService interface {
    GetFormOptions(ctx context.Context) (*schemas.FormOptionsResponse, error)
}

type formOptionsService struct {
    repo repositories.FormOptionsRepository
}

func NewFormOptionsService(repo repositories.FormOptionsRepository) FormOptionsService {
    return &formOptionsService{repo: repo}
}

func (s *formOptionsService) GetFormOptions(ctx context.Context) (*schemas.FormOptionsResponse, error) {
    services, err := s.repo.GetAllNested(ctx)
    if err != nil {
        return nil, err
    }

    docTypes, err := s.repo.GetDocumentTypes(ctx)
    if err != nil {
        return nil, err
    }

    return &schemas.FormOptionsResponse{
        Services:      services,
        DocumentTypes: docTypes,
    }, nil
}