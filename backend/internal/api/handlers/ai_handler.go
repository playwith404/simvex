package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"simvex/internal/models"
	"simvex/internal/services"
)

type AIHandler struct {
	service *services.AIService
}

func NewAIHandler(service *services.AIService) *AIHandler {
	return &AIHandler{service: service}
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
		return
	}

	resp, err := h.service.Chat(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidMessage) {
			respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
			return
		}
		if errors.Is(err, services.ErrInvalidID) {
			respondError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "요청한 정보를 찾을 수 없습니다", nil)
			return
		}
		respondError(c, http.StatusBadGateway, "AI_SERVICE_ERROR", "AI 서비스 오류가 발생했습니다", nil)
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{AssistantMessage: resp})
}
