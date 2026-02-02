package models

type ChatRequest struct {
	ObjectID    string    `json:"objectId"`
	PartID      string    `json:"partId,omitempty"`
	UserMessage string    `json:"userMessage"`
	History     []Message `json:"history,omitempty"`
}

type ChatResponse struct {
	AssistantMessage string `json:"assistantMessage"`
}

type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}
