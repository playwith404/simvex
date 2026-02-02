package models

import "time"

type Object struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Thumbnail   string    `json:"thumbnail"`
	ModelPath   string    `json:"modelPath"`
	Theory      string    `json:"theory"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"createdAt"`
}
