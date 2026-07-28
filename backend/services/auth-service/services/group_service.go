package services

import (
	"auth-service/models"
	"auth-service/repositories"
)

type GroupService struct {
	groupRepo *repositories.GroupRepository
}

func NewGroupService(groupRepo *repositories.GroupRepository) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

func (s *GroupService) GetAll() ([]models.Group, error) {
	return s.groupRepo.GetAll()
}

func (s *GroupService) Exists(groupID string) (bool, error) {
	return s.groupRepo.Exists(groupID)
}
