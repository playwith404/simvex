package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"simvex/internal/repository"
)

type NotionService struct {
	repo repository.Repository
	key  []byte
}

func NewNotionService(repo repository.Repository, keyBase64 string) (*NotionService, error) {
	if strings.TrimSpace(keyBase64) == "" {
		return &NotionService{repo: repo, key: nil}, nil
	}
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid NOTION_TOKEN_KEY: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("NOTION_TOKEN_KEY must be 32 bytes")
	}
	return &NotionService{repo: repo, key: key}, nil
}

func (s *NotionService) SetToken(ctx context.Context, userID, token string) error {
	if s.key == nil {
		return fmt.Errorf("notion token key not configured")
	}
	encrypted, err := encryptToken(s.key, token)
	if err != nil {
		return err
	}
	return s.repo.SetNotionToken(userID, encrypted)
}

func (s *NotionService) GetToken(ctx context.Context, userID string) (string, bool, error) {
	if s.key == nil {
		return "", false, fmt.Errorf("notion token key not configured")
	}
	enc, err := s.repo.GetNotionToken(userID)
	if err != nil {
		return "", false, err
	}
	if enc == "" {
		return "", false, nil
	}
	token, err := decryptToken(s.key, enc)
	if err != nil {
		return "", false, err
	}
	return token, true, nil
}

func (s *NotionService) DeleteToken(ctx context.Context, userID string) error {
	return s.repo.DeleteNotionToken(userID)
}

func encryptToken(key []byte, token string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptToken(key []byte, encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid token")
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
