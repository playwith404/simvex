package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simvex/internal/api/middleware"
	"simvex/internal/services"
)

type AuthHandler struct {
	service      *services.AuthService
	cookieName   string
	cookieSecure bool
}

func NewAuthHandler(service *services.AuthService, cookieName string, cookieSecure bool) *AuthHandler {
	return &AuthHandler{service: service, cookieName: cookieName, cookieSecure: cookieSecure}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type resetRequest struct {
	Email string `json:"email"`
}

type resetConfirmRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
		return
	}
	if err := h.service.Register(c.Request.Context(), req.Email, req.Password); err != nil {
		switch err {
		case services.ErrUserAlreadyExists:
			respondError(c, http.StatusConflict, "USER_EXISTS", "이미 가입된 이메일입니다", nil)
		case services.ErrInvalidCredentials:
			respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "이메일 또는 비밀번호가 유효하지 않습니다", nil)
		default:
			respondError(c, http.StatusInternalServerError, "REGISTER_FAILED", "회원가입에 실패했습니다", nil)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "verification_sent"})
}

func (h *AuthHandler) Verify(c *gin.Context) {
	var req verifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
		return
	}
	if err := h.service.VerifyEmail(c.Request.Context(), req.Email, req.Code); err != nil {
		switch err {
		case services.ErrCodeMismatch:
			respondError(c, http.StatusBadRequest, "CODE_MISMATCH", "인증 코드가 올바르지 않습니다", nil)
		case services.ErrCodeExpired:
			respondError(c, http.StatusBadRequest, "CODE_EXPIRED", "인증 코드가 만료되었습니다", nil)
		case services.ErrUserNotFound:
			respondError(c, http.StatusNotFound, "USER_NOT_FOUND", "계정을 찾을 수 없습니다", nil)
		default:
			respondError(c, http.StatusInternalServerError, "VERIFY_FAILED", "인증에 실패했습니다", nil)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "verified"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
		return
	}
	sessionID, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			respondError(c, http.StatusNotFound, "USER_NOT_FOUND", "계정을 찾을 수 없습니다", nil)
		case services.ErrUserNotVerified:
			respondError(c, http.StatusForbidden, "USER_NOT_VERIFIED", "이메일 인증이 필요합니다", nil)
		case services.ErrInvalidCredentials:
			respondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "이메일 또는 비밀번호가 올바르지 않습니다", nil)
		default:
			respondError(c, http.StatusInternalServerError, "LOGIN_FAILED", "로그인에 실패했습니다", nil)
		}
		return
	}

	c.SetCookie(h.cookieName, sessionID, int((14*24*time.Hour).Seconds()), "/", "", h.cookieSecure, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged_in"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, _ := c.Cookie(h.cookieName)
	_ = h.service.Logout(c.Request.Context(), sessionID)
	c.SetCookie(h.cookieName, "", -1, "/", "", h.cookieSecure, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged_out"})
}

func (h *AuthHandler) RequestReset(c *gin.Context) {
	var req resetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
		return
	}
	if err := h.service.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		respondError(c, http.StatusInternalServerError, "RESET_REQUEST_FAILED", "요청에 실패했습니다", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reset_sent"})
}

func (h *AuthHandler) ConfirmReset(c *gin.Context) {
	var req resetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "잘못된 요청입니다", nil)
		return
	}
	if err := h.service.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		switch err {
		case services.ErrCodeMismatch:
			respondError(c, http.StatusBadRequest, "CODE_MISMATCH", "코드가 올바르지 않습니다", nil)
		case services.ErrCodeExpired:
			respondError(c, http.StatusBadRequest, "CODE_EXPIRED", "코드가 만료되었습니다", nil)
		case services.ErrUserNotFound:
			respondError(c, http.StatusNotFound, "USER_NOT_FOUND", "계정을 찾을 수 없습니다", nil)
		case services.ErrInvalidCredentials:
			respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "비밀번호가 유효하지 않습니다", nil)
		default:
			respondError(c, http.StatusInternalServerError, "RESET_FAILED", "비밀번호 변경에 실패했습니다", nil)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password_updated"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "로그인이 필요합니다", nil)
		return
	}
	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "USER_LOAD_FAILED", "사용자 정보를 불러오지 못했습니다", nil)
		return
	}
	if user == nil {
		respondError(c, http.StatusNotFound, "USER_NOT_FOUND", "사용자를 찾을 수 없습니다", nil)
		return
	}
	c.JSON(http.StatusOK, user)
}
