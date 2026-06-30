package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"goplayground/backend/internal/repository"
)

type UsersHandler struct {
	userRepo *repository.UserRepository
}

func NewUsersHandler(userRepo *repository.UserRepository) *UsersHandler {
	return &UsersHandler{userRepo: userRepo}
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Fullname  string `json:"fullname"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type UpdateUserRequest struct {
	Fullname *string `json:"fullname"`
	Email    *string `json:"email"`
	Role     *string `json:"role"`
}

func (h *UsersHandler) List(c *gin.Context) {
	if getUserRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}
	users, err := h.userRepo.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]UserResponse, len(users))
	for i, u := range users {
		out[i] = UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			Fullname:  u.Fullname,
			Role:      u.Role,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04"),
		}
	}
	c.JSON(http.StatusOK, out)
}

func (h *UsersHandler) Update(c *gin.Context) {
	if getUserRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	user, err := h.userRepo.FindByID(uint(id))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Fullname != nil {
		user.Fullname = *req.Fullname
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		r := *req.Role
		if r == "admin" || r == "student" {
			user.Role = r
		}
	}
	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Fullname:  user.Fullname,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04"),
	})
}
