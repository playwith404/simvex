package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"simvex/internal/api/middleware"
	"simvex/internal/models"
	"simvex/internal/services"
)

type WorkflowHandler struct {
	service *services.WorkflowService
}

func NewWorkflowHandler(service *services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{service: service}
}

type projectRequest struct {
	Title string `json:"title"`
}

type fullRequest struct {
	Nodes       []models.WorkflowNode       `json:"nodes"`
	Edges       []models.WorkflowEdge       `json:"edges"`
	Checklists  []models.WorkflowChecklist  `json:"checklists"`
	Attachments []models.WorkflowAttachment `json:"attachments"`
}

func (h *WorkflowHandler) ListProjects(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	projects, err := h.service.ListProjects(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "PROJECT_LIST_FAILED", "프로젝트 조회에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (h *WorkflowHandler) CreateProject(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "프로젝트 제목이 필요합니다", nil)
		return
	}
	project, err := h.service.CreateProject(c.Request.Context(), userID, req.Title)
	if err != nil {
		if err == services.ErrInvalidTitle {
			respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "프로젝트 제목이 필요합니다", nil)
			return
		}
		respondError(c, http.StatusInternalServerError, "PROJECT_CREATE_FAILED", "프로젝트 생성에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *WorkflowHandler) UpdateProject(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	projectID := c.Param("id")
	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "프로젝트 제목이 필요합니다", nil)
		return
	}
	project, err := h.service.UpdateProject(c.Request.Context(), userID, projectID, req.Title)
	if err != nil {
		if err == services.ErrInvalidTitle {
			respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "프로젝트 제목이 필요합니다", nil)
			return
		}
		if err == services.ErrNotFound {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "프로젝트를 찾을 수 없습니다", nil)
			return
		}
		respondError(c, http.StatusInternalServerError, "PROJECT_UPDATE_FAILED", "프로젝트 수정에 실패했습니다", nil)
		return
	}
	if project == nil {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "프로젝트를 찾을 수 없습니다", nil)
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *WorkflowHandler) DeleteProject(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	projectID := c.Param("id")
	if err := h.service.DeleteProject(c.Request.Context(), userID, projectID); err != nil {
		respondError(c, http.StatusInternalServerError, "PROJECT_DELETE_FAILED", "프로젝트 삭제에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *WorkflowHandler) GetFull(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	projectID := c.Param("id")
	full, err := h.service.GetFull(c.Request.Context(), userID, projectID)
	if err != nil {
		if err == services.ErrNotFound {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "프로젝트를 찾을 수 없습니다", nil)
			return
		}
		respondError(c, http.StatusInternalServerError, "WORKFLOW_LOAD_FAILED", "워크플로우 조회에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, full)
}

func (h *WorkflowHandler) SaveFull(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	projectID := c.Param("id")
	var req fullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
		return
	}
	if err := h.service.SaveFull(c.Request.Context(), userID, projectID, req.Nodes, req.Edges, req.Checklists, req.Attachments); err != nil {
		respondError(c, http.StatusInternalServerError, "WORKFLOW_SAVE_FAILED", "워크플로우 저장에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "saved"})
}
