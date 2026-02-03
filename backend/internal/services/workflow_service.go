package services

import (
	"context"

	"simvex/internal/models"
	"simvex/internal/repository"
)

type WorkflowService struct {
	repo repository.Repository
}

func NewWorkflowService(repo repository.Repository) *WorkflowService {
	return &WorkflowService{repo: repo}
}

func (s *WorkflowService) CreateProject(ctx context.Context, userID, title string) (*models.WorkflowProject, error) {
	if title == "" {
		return nil, ErrInvalidTitle
	}
	return s.repo.CreateProject(userID, title)
}

func (s *WorkflowService) ListProjects(ctx context.Context, userID string) ([]models.WorkflowProject, error) {
	return s.repo.ListProjects(userID)
}

func (s *WorkflowService) UpdateProject(ctx context.Context, userID, projectID, title string) (*models.WorkflowProject, error) {
	if title == "" {
		return nil, ErrInvalidTitle
	}
	return s.repo.UpdateProject(userID, projectID, title)
}

func (s *WorkflowService) DeleteProject(ctx context.Context, userID, projectID string) error {
	return s.repo.DeleteProject(userID, projectID)
}

func (s *WorkflowService) GetFull(ctx context.Context, userID, projectID string) (*models.WorkflowFull, error) {
	project, err := s.repo.GetProject(userID, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrNotFound
	}
	nodes, edges, checklists, attachments, err := s.repo.LoadWorkflowFull(userID, projectID)
	if err != nil {
		return nil, err
	}
	return &models.WorkflowFull{
		Project:     *project,
		Nodes:       nodes,
		Edges:       edges,
		Checklists:  checklists,
		Attachments: attachments,
	}, nil
}

func (s *WorkflowService) SaveFull(ctx context.Context, userID, projectID string, nodes []models.WorkflowNode, edges []models.WorkflowEdge, checklists []models.WorkflowChecklist, attachments []models.WorkflowAttachment) error {
	return s.repo.SaveWorkflowFull(userID, projectID, nodes, edges, checklists, attachments)
}
