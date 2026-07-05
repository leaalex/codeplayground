package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"goplayground/backend/internal/runner"
)

type RunHandler struct {
	runner  *runner.DockerRunner
	timeout time.Duration
}

func NewRunHandler(r *runner.DockerRunner, timeout time.Duration) *RunHandler {
	return &RunHandler{runner: r, timeout: timeout}
}

type RunRequest struct {
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required,oneof=go python"`
}

type RunResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func (h *RunHandler) Run(c *gin.Context) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Code) > 100_000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code too large"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()
	lang, ok := runner.ParseLanguage(req.Language)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported language"})
		return
	}
	res := h.runner.Run(ctx, lang, req.Code)
	c.JSON(http.StatusOK, RunResponse{
		Output: res.Output,
		Error:  res.Error,
	})
}
