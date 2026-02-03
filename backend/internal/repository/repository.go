package repository

import "simvex/internal/models"

type Repository interface {
	Close() error
	GetObjects() ([]models.Object, error)
	GetObjectByID(id string) (*models.Object, error)
	GetPartsByObjectID(objectID string) ([]models.Part, error)
	GetPartByID(partID string) (*models.Part, error)
	GetAssetByPath(path string) (*models.Asset, error)
	CreateUser(email, passwordHash string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	SetUserVerified(id string) error
	UpdateUserPassword(id, passwordHash string) error
	CreateProject(userID, title string) (*models.WorkflowProject, error)
	ListProjects(userID string) ([]models.WorkflowProject, error)
	UpdateProject(userID, projectID, title string) (*models.WorkflowProject, error)
	DeleteProject(userID, projectID string) error
	GetProject(userID, projectID string) (*models.WorkflowProject, error)
	SaveWorkflowFull(userID, projectID string, nodes []models.WorkflowNode, edges []models.WorkflowEdge, checklists []models.WorkflowChecklist, attachments []models.WorkflowAttachment) error
	LoadWorkflowFull(userID, projectID string) ([]models.WorkflowNode, []models.WorkflowEdge, []models.WorkflowChecklist, []models.WorkflowAttachment, error)
	GetNoteByPart(userID, partID string) (*models.Note, error)
	UpsertNote(userID, partID, content string) (*models.Note, error)
	SetNotionToken(userID, tokenEncrypted string) error
	GetNotionToken(userID string) (string, error)
	DeleteNotionToken(userID string) error
}
