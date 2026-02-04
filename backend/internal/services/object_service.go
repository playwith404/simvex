package services

import (
	"simvex/internal/models"
	"simvex/internal/repository"
)

type ObjectService struct {
	repo repository.Repository
}

func NewObjectService(repo repository.Repository) *ObjectService {
	return &ObjectService{repo: repo}
}

func (s *ObjectService) GetObjects() ([]models.Object, error) {
	return s.repo.GetObjects()
}

func (s *ObjectService) GetObjectByID(id string) (*models.Object, error) {
	return s.repo.GetObjectByID(id)
}

func (s *ObjectService) GetPartsByObjectID(id string) ([]models.Part, error) {
	return s.repo.GetPartsByObjectID(id)
}

func (s *ObjectService) GetPartByID(id string) (*models.Part, error) {
	return s.repo.GetPartByID(id)
}

func (s *ObjectService) GetObjectVersions(objectID string) ([]models.Object, error) {
	return s.repo.GetObjectVersions(objectID)
}
