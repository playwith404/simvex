package models

import "time"

type WorkflowProject struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Title        string    `json:"title"`
	NotionPageID string    `json:"notionPageId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type WorkflowNode struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"projectId"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	ScheduledDate string   `json:"scheduledDate"`
	Progress     int       `json:"progress"`
	Color        string    `json:"color,omitempty"`
	PositionX    float64   `json:"positionX"`
	PositionY    float64   `json:"positionY"`
	LinkedPartID string    `json:"linkedPartId,omitempty"`
	LinkedNoteID string    `json:"linkedNoteId,omitempty"`
	NotionPageID string    `json:"notionPageId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type WorkflowEdge struct {
	ID         string `json:"id"`
	ProjectID  string `json:"projectId"`
	SourceID   string `json:"source"`
	TargetID   string `json:"target"`
}

type WorkflowChecklist struct {
	ID     string `json:"id"`
	NodeID string `json:"nodeId"`
	Text   string `json:"text"`
	Done   bool   `json:"done"`
}

type WorkflowAttachment struct {
	ID     string `json:"id"`
	NodeID string `json:"nodeId"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	URL    string `json:"url"`
}

type WorkflowFull struct {
	Project    WorkflowProject      `json:"project"`
	Nodes      []WorkflowNode       `json:"nodes"`
	Edges      []WorkflowEdge       `json:"edges"`
	Checklists []WorkflowChecklist  `json:"checklists"`
	Attachments []WorkflowAttachment `json:"attachments"`
}
