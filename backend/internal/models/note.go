package models

import "time"

type Note struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	PartID       string    `json:"partId"`
	Content      string    `json:"content"`
	NotionPageID string    `json:"notionPageId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
