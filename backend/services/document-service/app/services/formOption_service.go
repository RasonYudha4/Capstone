package services

import (
	"capstone/app/repositories"
	"capstone/app/schemas"
	"context"
)

type FormOptionsService interface {
	GetFormOptions(ctx context.Context) (*schemas.FormOptionsResponse, error)
	GetFormOptionsByGroup(ctx context.Context, userId string) (*schemas.FormOptionsResponse, error)
}

type formOptionsService struct {
	repo repositories.FormOptionsRepository
}

func NewFormOptionsService(repo repositories.FormOptionsRepository) FormOptionsService {
	return &formOptionsService{repo: repo}
}

// GetFormOptions returns all form options (used for breadcrumbs / navigation).
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

// GetFormOptionsByGroup returns form options filtered by the user's group (used for upload form).
func (s *formOptionsService) GetFormOptionsByGroup(ctx context.Context, userId string) (*schemas.FormOptionsResponse, error) {
	services, err := s.repo.GetAllNestedByGroup(ctx, userId)
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
