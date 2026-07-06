package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"goplayground/backend/internal/presence"
	"goplayground/backend/internal/repository"
)

type PresenceHandler struct {
	store    *presence.Store
	fileRepo *repository.FileRepository
	userRepo *repository.UserRepository
}

func NewPresenceHandler(store *presence.Store, fileRepo *repository.FileRepository, userRepo *repository.UserRepository) *PresenceHandler {
	return &PresenceHandler{store: store, fileRepo: fileRepo, userRepo: userRepo}
}

type presenceResponse struct {
	UserID   uint   `json:"user_id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
}

func (h *PresenceHandler) ListAll(c *gin.Context) {
	if getUserRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}
	sessions := h.store.GetAll()
	out := make(map[string]presenceResponse, len(sessions))
	for fileID, session := range sessions {
		out[strconv.FormatUint(uint64(fileID), 10)] = presenceResponse{
			UserID:   session.UserID,
			Fullname: session.Fullname,
			Email:    session.Email,
		}
	}
	c.JSON(http.StatusOK, out)
}

func (h *PresenceHandler) Touch(c *gin.Context) {
	userID := getUserID(c)
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	file, err := h.fileRepo.GetByID(uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	if file.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only file owner can register presence"})
		return
	}
	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	h.store.Touch(uint(fileID), presence.UserInfo{
		UserID:   user.ID,
		Fullname: user.Fullname,
		Email:    user.Email,
	})
	c.Status(http.StatusNoContent)
}

func (h *PresenceHandler) Leave(c *gin.Context) {
	userID := getUserID(c)
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	h.store.Leave(uint(fileID), userID)
	c.Status(http.StatusNoContent)
}
