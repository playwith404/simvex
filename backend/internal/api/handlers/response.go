package handlers

import "github.com/gin-gonic/gin"

type errorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

func respondError(c *gin.Context, status int, code string, message string, details interface{}) {
	c.JSON(status, errorResponse{Error: errorDetail{Code: code, Message: message, Details: details}})
}
