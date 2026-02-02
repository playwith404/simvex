package services

import (
	"simvex/internal/models"
	"simvex/internal/repository"
)

type AssetService struct {
	repo repository.Repository
}

func NewAssetService(repo repository.Repository) *AssetService {
	return &AssetService{repo: repo}
}

func (s *AssetService) GetAsset(path string) (*models.Asset, error) {
	return s.repo.GetAssetByPath(path)
}
