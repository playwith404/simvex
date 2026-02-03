package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"simvex/internal/api/middleware"
	"simvex/internal/services"
)

type NoteHandler struct {
	service *services.NoteService
}

func NewNoteHandler(service *services.NoteService) *NoteHandler {
	return &NoteHandler{service: service}
}

type noteRequest struct {
	Content string `json:"content"`
}

func (h *NoteHandler) GetNote(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	partID := c.Param("partId")
	note, err := h.service.GetNote(c.Request.Context(), userID, partID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "NOTE_LOAD_FAILED", "노트 조회에 실패했습니다", nil)
		return
	}
	if note == nil {
		c.JSON(http.StatusOK, gin.H{"content": ""})
		return
	}
	c.JSON(http.StatusOK, note)
}

func (h *NoteHandler) UpsertNote(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	partID := c.Param("partId")
	var req noteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
		return
	}
	note, err := h.service.UpsertNote(c.Request.Context(), userID, partID, req.Content)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "NOTE_SAVE_FAILED", "노트 저장에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, note)
}
