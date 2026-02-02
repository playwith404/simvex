package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"simvex/internal/services"
)

type PartHandler struct {
	service *services.ObjectService
}

func NewPartHandler(service *services.ObjectService) *PartHandler {
	return &PartHandler{service: service}
}

func (h *PartHandler) GetPart(c *gin.Context) {
	partID := c.Param("partId")
	part, err := h.service.GetPartByID(partID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 오류가 발생했습니다", nil)
		return
	}
	if part == nil {
		respondError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "요청한 부품을 찾을 수 없습니다", gin.H{"partId": partID})
		return
	}
	c.JSON(http.StatusOK, part)
}
