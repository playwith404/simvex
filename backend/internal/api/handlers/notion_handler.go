package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"simvex/internal/api/middleware"
	"simvex/internal/services"
)

type NotionHandler struct {
	service *services.NotionService
}

func NewNotionHandler(service *services.NotionService) *NotionHandler {
	return &NotionHandler{service: service}
}

type notionConnectRequest struct {
	Token string `json:"token"`
}

func (h *NotionHandler) Connect(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	var req notionConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "토큰이 필요합니다", nil)
		return
	}
	if err := h.service.SetToken(c.Request.Context(), userID, req.Token); err != nil {
		respondError(c, http.StatusInternalServerError, "NOTION_CONNECT_FAILED", "Notion 연결에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "connected"})
}

func (h *NotionHandler) Status(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	_, okToken, err := h.service.GetToken(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "NOTION_STATUS_FAILED", "Notion 상태 조회에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"connected": okToken})
}

func (h *NotionHandler) Disconnect(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	if err := h.service.DeleteToken(c.Request.Context(), userID); err != nil {
		respondError(c, http.StatusInternalServerError, "NOTION_DISCONNECT_FAILED", "Notion 연결 해제에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "disconnected"})
}

func (h *NotionHandler) Sync(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	_, okToken, err := h.service.GetToken(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "NOTION_SYNC_FAILED", "동기화에 실패했습니다", nil)
		return
	}
	if !okToken {
		respondError(c, http.StatusBadRequest, "NOTION_NOT_CONNECTED", "Notion 토큰이 필요합니다", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sync_started"})
}
