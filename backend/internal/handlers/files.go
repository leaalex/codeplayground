package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"goplayground/backend/internal/models"
	"goplayground/backend/internal/repository"
)

type FilesHandler struct {
	fileRepo *repository.FileRepository
	userRepo *repository.UserRepository
}

func NewFilesHandler(fileRepo *repository.FileRepository, userRepo *repository.UserRepository) *FilesHandler {
	return &FilesHandler{fileRepo: fileRepo, userRepo: userRepo}
}

type CreateFileRequest struct {
	Name    string `json:"name" binding:"required"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type UpdateFileRequest struct {
	Name     *string `json:"name"`
	Path     *string `json:"path"`
	Content  *string `json:"content"`
	Verified        *bool   `json:"verified"`
	UserID          *uint   `json:"user_id"`
	AutosaveEnabled *bool   `json:"autosave_enabled"`
}

func (h *FilesHandler) List(c *gin.Context) {
	userID := getUserID(c)
	role := getUserRole(c)
	var files []models.File
	var err error
	if role == "admin" {
		files, err = h.fileRepo.ListAll()
	} else {
		files, err = h.fileRepo.List(userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

func (h *FilesHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req CreateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	file := &models.File{
		UserID:          userID,
		Name:            req.Name,
		Path:            req.Path,
		Content:         req.Content,
		AutosaveEnabled: true,
	}
	if err := h.fileRepo.Create(file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, file)
}

func (h *FilesHandler) Get(c *gin.Context) {
	userID := getUserID(c)
	role := getUserRole(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var file *models.File
	if role == "admin" {
		file, err = h.fileRepo.GetByID(uint(id))
	} else {
		file, err = h.fileRepo.Get(uint(id), userID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	c.JSON(http.StatusOK, file)
}

func (h *FilesHandler) Update(c *gin.Context) {
	userID := getUserID(c)
	role := getUserRole(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var file *models.File
	if role == "admin" {
		file, err = h.fileRepo.GetByID(uint(id))
	} else {
		file, err = h.fileRepo.Get(uint(id), userID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	var req UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != nil {
		file.Name = *req.Name
	}
	if req.Path != nil {
		file.Path = *req.Path
	}
	if req.Content != nil {
		file.Content = *req.Content
	}
	if req.Verified != nil && role == "admin" {
		file.Verified = *req.Verified
	}
	if req.UserID != nil && role == "admin" && *req.UserID != file.UserID {
		owner, err := h.userRepo.FindByID(*req.UserID)
		if err != nil || owner == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}
		file.UserID = *req.UserID
	}
	if req.AutosaveEnabled != nil && role == "admin" {
		file.AutosaveEnabled = *req.AutosaveEnabled
	}
	if err := h.fileRepo.Update(file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role == "admin" {
		file, err = h.fileRepo.GetByID(file.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, file)
}

func (h *FilesHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	role := getUserRole(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if role == "admin" {
		err = h.fileRepo.DeleteByID(uint(id))
	} else {
		err = h.fileRepo.Delete(uint(id), userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func getUserID(c *gin.Context) uint {
	v, _ := c.Get("userID")
	switch id := v.(type) {
	case float64:
		return uint(id)
	case uint:
		return id
	default:
		return 0
	}
}

func getUserRole(c *gin.Context) string {
	v, _ := c.Get("userRole")
	if role, ok := v.(string); ok && role == "admin" {
		return "admin"
	}
	return "student"
}
