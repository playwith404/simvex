package services

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"simvex/internal/models"
	"simvex/internal/repository"
)

type NotionService struct {
	repo repository.Repository
	key  []byte
	http *http.Client
}

func NewNotionService(repo repository.Repository, keyBase64 string) (*NotionService, error) {
	client := &http.Client{Timeout: 45 * time.Second}
	if strings.TrimSpace(keyBase64) == "" {
		return &NotionService{repo: repo, key: nil, http: client}, nil
	}
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid NOTION_TOKEN_KEY: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("NOTION_TOKEN_KEY must be 32 bytes")
	}
	return &NotionService{
		repo: repo,
		key:  key,
		http: client,
	}, nil
}

func (s *NotionService) SetToken(ctx context.Context, userID, token, parentPageID string) error {
	if s.key == nil {
		return fmt.Errorf("notion token key not configured")
	}
	if strings.TrimSpace(parentPageID) == "" {
		return fmt.Errorf("parent page ID is required")
	}
	encrypted, err := encryptToken(s.key, token)
	if err != nil {
		return err
	}
	return s.repo.SetNotionToken(userID, encrypted, parentPageID)
}

func (s *NotionService) GetToken(ctx context.Context, userID string) (token string, parentPageID string, ok bool, err error) {
	if s.key == nil {
		return "", "", false, nil
	}
	enc, parentPageID, err := s.repo.GetNotionToken(userID)
	if err != nil {
		return "", "", false, err
	}
	if enc == "" {
		return "", "", false, nil
	}
	token, err = decryptToken(s.key, enc)
	if err != nil {
		return "", "", false, err
	}
	return token, parentPageID, true, nil
}

func (s *NotionService) DeleteToken(ctx context.Context, userID string) error {
	return s.repo.DeleteNotionToken(userID)
}

func (s *NotionService) SyncUser(ctx context.Context, userID string) error {
	token, parentPageID, ok, err := s.GetToken(ctx, userID)
	if err != nil {
		return err
	}
	if !ok || token == "" {
		return fmt.Errorf("notion token not connected")
	}
	if strings.TrimSpace(parentPageID) == "" {
		return fmt.Errorf("notion parent page ID not configured")
	}

	projects, err := s.repo.ListProjects(userID)
	if err != nil {
		return err
	}

	for _, project := range projects {
		projectPageID := project.NotionPageID
		if projectPageID == "" {
			pageID, err := s.createProjectPage(ctx, token, parentPageID, project.Title)
			if err != nil {
				return err
			}
			projectPageID = pageID
			if err := s.repo.UpdateProjectNotionPageID(userID, project.ID, pageID); err != nil {
				return err
			}
		} else {
			if err := s.updatePageTitle(ctx, token, projectPageID, project.Title); err != nil {
				return err
			}
		}

		nodes, _, checklists, attachments, err := s.repo.LoadWorkflowFull(userID, project.ID)
		if err != nil {
			return err
		}

		objectNames := make(map[string]string)
		objectNotes := make(map[string]string)
		for _, node := range nodes {
			objectID := strings.TrimSpace(node.LinkedPartID)
			if objectID == "" {
				continue
			}
			if _, ok := objectNames[objectID]; !ok {
				obj, err := s.repo.GetObjectByID(objectID)
				if err != nil {
					return err
				}
				if obj != nil {
					objectNames[objectID] = obj.Name
				}
			}
			if _, ok := objectNotes[objectID]; !ok {
				note, err := s.repo.GetNoteByPart(userID, objectID)
				if err != nil {
					return err
				}
				if note != nil {
					objectNotes[objectID] = note.Content
				}
			}
		}

		keepPages := make(map[string]struct{}, len(nodes))
		for _, node := range nodes {
			if node.NotionPageID != "" {
				keepPages[node.NotionPageID] = struct{}{}
			}
			objectName := objectNames[strings.TrimSpace(node.LinkedPartID)]
			objectNote := objectNotes[strings.TrimSpace(node.LinkedPartID)]
			if node.NotionPageID == "" {
				pageID, err := s.createNodePage(ctx, token, projectPageID, node, checklists, attachments, objectName, objectNote)
				if err != nil {
					return err
				}
				keepPages[pageID] = struct{}{}
				if err := s.repo.UpdateNodeNotionPageID(userID, project.ID, node.ID, pageID); err != nil {
					return err
				}
			} else {
				if err := s.updateNodeProperties(ctx, token, node, checklists, attachments, objectName, objectNote); err != nil {
					return err
				}
			}
		}
if err := s.cleanupOrphanNodePages(ctx, token, projectPageID, keepPages); err != nil {
			return err
		}
	}

	return nil
}

type notionCreatePageResponse struct {
	ID string `json:"id"`
}

func (s *NotionService) createProjectPage(ctx context.Context, token, parentPageID, title string) (string, error) {
	if strings.TrimSpace(parentPageID) == "" {
		return "", fmt.Errorf("parent page ID is required")
	}
	parent := map[string]interface{}{
		"page_id": parentPageID,
	}
	payload := map[string]interface{}{
		"parent": parent,
		"properties": map[string]interface{}{
			"title": []map[string]interface{}{
				{"text": map[string]interface{}{"content": title}},
			},
		},
	}
	var resp notionCreatePageResponse
	if err := s.callNotion(ctx, token, "POST", "https://api.notion.com/v1/pages", payload, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (s *NotionService) createNodePage(ctx context.Context, token, parentID string, node models.WorkflowNode, checklists []models.WorkflowChecklist, attachments []models.WorkflowAttachment, objectName, objectNote string) (string, error) {
	children := buildNodeChildren(node, checklists, attachments, objectName, objectNote)
	props := map[string]interface{}{
		"title": []map[string]interface{}{
			{"text": map[string]interface{}{"content": node.Title}},
		},
	}
	payload := map[string]interface{}{
		"parent": map[string]interface{}{
			"page_id": parentID,
		},
		"properties": props,
	}
	if len(children) > 0 {
		payload["children"] = children
	}
	var resp notionCreatePageResponse
	if err := s.callNotion(ctx, token, "POST", "https://api.notion.com/v1/pages", payload, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (s *NotionService) updatePageTitle(ctx context.Context, token, pageID, title string) error {
	payload := map[string]interface{}{
		"properties": map[string]interface{}{
			"title": []map[string]interface{}{
				{"text": map[string]interface{}{"content": title}},
			},
		},
	}
	return s.callNotion(ctx, token, "PATCH", "https://api.notion.com/v1/pages/"+pageID, payload, nil)
}

func (s *NotionService) updateNodeProperties(ctx context.Context, token string, node models.WorkflowNode, checklists []models.WorkflowChecklist, attachments []models.WorkflowAttachment, objectName, objectNote string) error {
	props := map[string]interface{}{
		"title": []map[string]interface{}{
			{"text": map[string]interface{}{"content": node.Title}},
		},
	}
	payload := map[string]interface{}{
		"properties": props,
	}
	if err := s.callNotion(ctx, token, "PATCH", "https://api.notion.com/v1/pages/"+node.NotionPageID, payload, nil); err != nil {
		return err
	}
	return s.replaceNodeChildren(ctx, token, node.NotionPageID, node, checklists, attachments, objectName, objectNote)
}

func buildNodeChildren(node models.WorkflowNode, checklists []models.WorkflowChecklist, attachments []models.WorkflowAttachment, objectName, objectNote string) []map[string]interface{} {
	children := []map[string]interface{}{}
	children = append(children, map[string]interface{}{
		"object": "block",
		"type":   "heading_3",
		"heading_3": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"type": "text", "text": map[string]interface{}{"content": "SIMVEX_START"}},
			},
		},
	})
	children = append(children, map[string]interface{}{
		"object": "block",
		"type":   "paragraph",
		"paragraph": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"type": "text", "text": map[string]interface{}{"content": fmt.Sprintf("Scheduled: %s", node.ScheduledDate)}},
			},
		},
	})
	children = append(children, map[string]interface{}{
		"object": "block",
		"type":   "paragraph",
		"paragraph": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"type": "text", "text": map[string]interface{}{"content": fmt.Sprintf("Progress: %d", node.Progress)}},
			},
		},
	})
	if node.Color != "" {
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]interface{}{"content": fmt.Sprintf("Color: %s", node.Color)}},
				},
			},
		})
	}
	if node.Description != "" {
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]interface{}{"content": node.Description}},
				},
			},
		})
	}
	if strings.TrimSpace(node.LinkedPartID) != "" {
		label := objectName
		if strings.TrimSpace(label) == "" {
			label = node.LinkedPartID
		}
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]interface{}{"content": fmt.Sprintf("Object: %s", label)}},
				},
			},
		})
	}
	if strings.TrimSpace(objectNote) != "" {
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]interface{}{"content": "Object Note:"}},
				},
			},
		})
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]interface{}{"content": objectNote}},
				},
			},
		})
	}
	for _, c := range checklists {
		if c.NodeID != node.ID {
			continue
		}
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "to_do",
			"to_do": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]interface{}{"content": c.Text}},
				},
				"checked": c.Done,
			},
		})
	}
	for _, a := range attachments {
		if a.NodeID != node.ID {
			continue
		}
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]interface{}{"content": fmt.Sprintf("%s: %s", a.Name, a.URL)}},
				},
			},
		})
	}
	children = append(children, map[string]interface{}{
		"object": "block",
		"type":   "heading_3",
		"heading_3": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"type": "text", "text": map[string]interface{}{"content": "SIMVEX_END"}},
			},
		},
	})
	return children
}

type notionBlock struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Heading3 *struct {
		RichText []struct {
			Text struct {
				Content string `json:"content"`
			} `json:"text"`
		} `json:"rich_text"`
	} `json:"heading_3,omitempty"`
	ChildPage *struct {
		Title string `json:"title"`
	} `json:"child_page,omitempty"`
}

type notionChildrenResponse struct {
	Results    []notionBlock `json:"results"`
	HasMore    bool          `json:"has_more"`
	NextCursor string        `json:"next_cursor"`
}

func (s *NotionService) replaceNodeChildren(ctx context.Context, token, pageID string, node models.WorkflowNode, checklists []models.WorkflowChecklist, attachments []models.WorkflowAttachment, objectName, objectNote string) error {
	children := buildNodeChildren(node, checklists, attachments, objectName, objectNote)
	blocks, err := s.listBlockChildren(ctx, token, pageID)
	if err != nil {
		return err
	}
	startDelete := false
	for _, block := range blocks {
		if block.Type == "heading_3" && block.Heading3 != nil && len(block.Heading3.RichText) > 0 {
			text := block.Heading3.RichText[0].Text.Content
			if text == "SIMVEX_START" {
				startDelete = true
			}
			if text == "SIMVEX_END" && startDelete {
				if err := s.archiveBlock(ctx, token, block.ID); err != nil {
					return err
				}
				startDelete = false
				continue
			}
		}
		if startDelete {
			if err := s.archiveBlock(ctx, token, block.ID); err != nil {
				return err
			}
		}
	}
	if len(children) == 0 {
		return nil
	}
	payload := map[string]interface{}{
		"children": children,
	}
	return s.callNotion(ctx, token, "PATCH", "https://api.notion.com/v1/blocks/"+pageID+"/children", payload, nil)
}

func (s *NotionService) listBlockChildren(ctx context.Context, token, blockID string) ([]notionBlock, error) {
	var all []notionBlock
	cursor := ""
	for {
		url := "https://api.notion.com/v1/blocks/" + blockID + "/children"
		if cursor != "" {
			url += "?start_cursor=" + cursor
		}
		var resp notionChildrenResponse
		if err := s.callNotion(ctx, token, "GET", url, nil, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Results...)
		if !resp.HasMore {
			break
		}
		cursor = resp.NextCursor
	}
	return all, nil
}

func (s *NotionService) archiveBlock(ctx context.Context, token, blockID string) error {
	payload := map[string]interface{}{
		"archived": true,
	}
	return s.callNotion(ctx, token, "PATCH", "https://api.notion.com/v1/blocks/"+blockID, payload, nil)
}

func (s *NotionService) cleanupOrphanNodePages(ctx context.Context, token, projectPageID string, keep map[string]struct{}) error {
	blocks, err := s.listBlockChildren(ctx, token, projectPageID)
	if err != nil {
		return err
	}
	for _, block := range blocks {
		if block.Type != "child_page" {
			continue
		}
		if _, ok := keep[block.ID]; ok {
			continue
		}
		isManaged, err := s.pageHasSimvexMarker(ctx, token, block.ID)
		if err != nil {
			return err
		}
		if !isManaged {
			continue
		}
		if err := s.archiveBlock(ctx, token, block.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *NotionService) pageHasSimvexMarker(ctx context.Context, token, pageID string) (bool, error) {
	blocks, err := s.listBlockChildren(ctx, token, pageID)
	if err != nil {
		return false, err
	}
	for _, block := range blocks {
		if block.Type != "heading_3" || block.Heading3 == nil || len(block.Heading3.RichText) == 0 {
			continue
		}
		if block.Heading3.RichText[0].Text.Content == "SIMVEX_START" {
			return true, nil
		}
	}
	return false, nil
}

func (s *NotionService) callNotion(ctx context.Context, token, method, url string, payload interface{}, out interface{}) error {
	const maxAttempts = 4
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var body io.Reader
		if payload != nil {
			raw, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			body = bytes.NewReader(raw)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Notion-Version", "2022-06-28")
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.http.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			b, _ := io.ReadAll(resp.Body)
			retryAfter := resp.Header.Get("Retry-After")
			resp.Body.Close()

			if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) && attempt < maxAttempts-1 {
				delay := time.Duration(500*(1<<attempt)) * time.Millisecond
				if retryAfter != "" {
					if sec, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && sec > 0 {
						delay = time.Duration(sec) * time.Second
					}
				}
				time.Sleep(delay)
				continue
			}
			return fmt.Errorf("notion api error: %s", string(b))
		}
		defer resp.Body.Close()
		if out != nil {
			return json.NewDecoder(resp.Body).Decode(out)
		}
		return nil
	}
	return fmt.Errorf("notion api error: exceeded retry attempts")
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
