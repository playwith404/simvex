package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"simvex/internal/services"
)

type AssetHandler struct {
	service *services.AssetService
}

func NewAssetHandler(service *services.AssetService) *AssetHandler {
	return &AssetHandler{service: service}
}

func (h *AssetHandler) GetAsset(c *gin.Context) {
	filepath := c.Param("filepath")
	if filepath == "" {
		respondError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "요청한 파일을 찾을 수 없습니다", nil)
		return
	}
	filepath = strings.TrimPrefix(filepath, "/")
	path := "/assets/models/" + filepath

	asset, err := h.service.GetAsset(path)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 오류가 발생했습니다", nil)
		return
	}
	if asset == nil {
		respondError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "요청한 파일을 찾을 수 없습니다", gin.H{"path": path})
		return
	}

	c.Header("Content-Type", asset.ContentType)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Data(http.StatusOK, asset.ContentType, asset.Data)
}
