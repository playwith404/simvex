package openai

import (
	"context"

	goopenai "github.com/sashabaranov/go-openai"
	"simvex/internal/models"
)

type Client struct {
	client *goopenai.Client
	model  string
}

func NewClient(apiKey string, model string) *Client {
	cfg := goopenai.DefaultConfig(apiKey)
	return &Client{client: goopenai.NewClientWithConfig(cfg), model: model}
}

func (c *Client) Chat(ctx context.Context, systemPrompt string, history []models.Message, userMessage string) (string, error) {
	messages := make([]goopenai.ChatCompletionMessage, 0, len(history)+2)
	messages = append(messages, goopenai.ChatCompletionMessage{
		Role:    goopenai.ChatMessageRoleSystem,
		Content: systemPrompt,
	})
	for _, msg := range history {
		role := goopenai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = goopenai.ChatMessageRoleAssistant
		}
		messages = append(messages, goopenai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}
	messages = append(messages, goopenai.ChatCompletionMessage{
		Role:    goopenai.ChatMessageRoleUser,
		Content: userMessage,
	})

	resp, err := c.client.CreateChatCompletion(ctx, goopenai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Temperature: 0.4,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}
