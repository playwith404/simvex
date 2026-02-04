package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"simvex/internal/models"
	"simvex/internal/services"
)

type ObjectHandler struct {
	service *services.ObjectService
}

func NewObjectHandler(service *services.ObjectService) *ObjectHandler {
	return &ObjectHandler{service: service}
}

func (h *ObjectHandler) ListObjects(c *gin.Context) {
	objects, err := h.service.GetObjects()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 오류가 발생했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, objects)
}

func (h *ObjectHandler) GetObject(c *gin.Context) {
	id := c.Param("id")
	obj, err := h.service.GetObjectByID(id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 오류가 발생했습니다", nil)
		return
	}
	if obj == nil {
		respondError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "요청한 오브젝트를 찾을 수 없습니다", gin.H{"objectId": id})
		return
	}
	c.JSON(http.StatusOK, obj)
}

func (h *ObjectHandler) GetObjectModel(c *gin.Context) {
	id := c.Param("id")
	obj, err := h.service.GetObjectByID(id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 오류가 발생했습니다", nil)
		return
	}
	if obj == nil || obj.ModelPath == "" {
		respondError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "요청한 모델 파일을 찾을 수 없습니다", gin.H{"objectId": id})
		return
	}
	c.File("." + obj.ModelPath)
}

func (h *ObjectHandler) GetPartsByObject(c *gin.Context) {
	id := c.Param("id")
	parts, err := h.service.GetPartsByObjectID(id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 오류가 발생했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, parts)
}

func (h *ObjectHandler) GetVersions(c *gin.Context) {
	id := c.Param("id")
	versions, err := h.service.GetObjectVersions(id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 오류가 발생했습니다", nil)
		return
	}
	if versions == nil {
		versions = []models.Object{}
	}
	c.JSON(http.StatusOK, versions)
}
