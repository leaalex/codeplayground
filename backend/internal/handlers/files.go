package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"goplayground/backend/internal/filekind"
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
	Name                 *string `json:"name"`
	Path                 *string `json:"path"`
	Content              *string `json:"content"`
	Verified             *bool   `json:"verified"`
	UserID               *uint   `json:"user_id"`
	AutosaveEnabled      *bool   `json:"autosave_enabled"`
	InstructionsFileID   *uint   `json:"instructions_file_id"`
	ClearInstructions    *bool   `json:"clear_instructions"`
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
	c.JSON(http.StatusOK, filesToListResponse(files))
}

func (h *FilesHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	role := getUserRole(c)
	var req CreateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := filekind.ValidateName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if filekind.IsMarkdown(req.Name) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can create markdown files"})
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
	c.JSON(http.StatusCreated, fileToResponse(file, true))
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
	c.JSON(http.StatusOK, fileToResponse(file, true))
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
		if err := filekind.ValidateName(*req.Name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if filekind.IsMarkdown(*req.Name) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can use markdown file names"})
			return
		}
		if filekind.IsMarkdown(file.Name) && !filekind.IsMarkdown(*req.Name) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot rename markdown file to code extension"})
			return
		}
		if filekind.IsCode(file.Name) && filekind.IsMarkdown(*req.Name) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot rename code file to markdown extension"})
			return
		}
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
		if filekind.IsMarkdown(file.Name) && owner.Role != "admin" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "markdown instructions cannot be owned by students; transfer the linked code file (.go/.py) instead",
			})
			return
		}
		file.UserID = *req.UserID
	}
	if req.AutosaveEnabled != nil && role == "admin" {
		file.AutosaveEnabled = *req.AutosaveEnabled
	}
	if role == "admin" && (req.InstructionsFileID != nil || (req.ClearInstructions != nil && *req.ClearInstructions)) {
		if !filekind.IsCode(file.Name) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "instructions can only be linked to code files"})
			return
		}
		if req.ClearInstructions != nil && *req.ClearInstructions {
			file.InstructionsFileID = nil
		} else if req.InstructionsFileID != nil {
			if *req.InstructionsFileID == 0 {
				file.InstructionsFileID = nil
			} else {
				mdFile, err := h.fileRepo.GetByID(*req.InstructionsFileID)
				if err != nil || !filekind.IsMarkdown(mdFile.Name) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "instructions file not found or not markdown"})
					return
				}
				file.InstructionsFileID = req.InstructionsFileID
			}
		}
	}
	if err := h.fileRepo.Update(file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	file, err = h.fileRepo.GetByID(file.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fileToResponse(file, true))
}

func (h *FilesHandler) Delete(c *gin.Context) {
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
	if filekind.IsMarkdown(file.Name) {
		count, err := h.fileRepo.CountInstructionsReferences(file.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("markdown file is linked to %d code file(s); unlink first", count),
			})
			return
		}
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
