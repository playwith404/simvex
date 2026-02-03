package handlers

import "github.com/gin-gonic/gin"

type errorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
}

func respondError(c *gin.Context, status int, code string, message string, detail interface{}) {
	c.JSON(status, errorResponse{Code: code, Message: message, Detail: detail})
}
