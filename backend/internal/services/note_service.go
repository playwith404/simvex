package services

import (
	"context"

	"simvex/internal/models"
	"simvex/internal/repository"
)

type NoteService struct {
	repo repository.Repository
}

func NewNoteService(repo repository.Repository) *NoteService {
	return &NoteService{repo: repo}
}

func (s *NoteService) GetNote(ctx context.Context, userID, partID string) (*models.Note, error) {
	if partID == "" {
		return nil, ErrNotFound
	}
	return s.repo.GetNoteByPart(userID, partID)
}

func (s *NoteService) UpsertNote(ctx context.Context, userID, partID, content string) (*models.Note, error) {
	if partID == "" {
		return nil, ErrNotFound
	}
	return s.repo.UpsertNote(userID, partID, content)
}
