package repository

import "simvex/internal/models"

type Repository interface {
	Close() error
	GetObjects() ([]models.Object, error)
	GetObjectByID(id string) (*models.Object, error)
	GetPartsByObjectID(objectID string) ([]models.Part, error)
	GetPartByID(partID string) (*models.Part, error)
	GetAssetByPath(path string) (*models.Asset, error)
}
