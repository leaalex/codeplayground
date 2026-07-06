package handlers

import "goplayground/backend/internal/models"

type instructionsSummary struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type fileResponse struct {
	ID                   uint                 `json:"id"`
	UserID               uint                 `json:"user_id"`
	User                 *models.User         `json:"user,omitempty"`
	Name                 string               `json:"name"`
	Path                 string               `json:"path"`
	Content              string               `json:"content"`
	Verified             bool                 `json:"verified"`
	AutosaveEnabled      bool                 `json:"autosave_enabled"`
	InstructionsFileID   *uint                `json:"instructions_file_id"`
	Instructions         *instructionsSummary `json:"instructions,omitempty"`
	CreatedAt            interface{}          `json:"created_at"`
	UpdatedAt            interface{}          `json:"updated_at"`
}

func fileToResponse(f *models.File, includeContent bool) fileResponse {
	resp := fileResponse{
		ID:                 f.ID,
		UserID:             f.UserID,
		User:               f.User,
		Name:               f.Name,
		Path:               f.Path,
		Verified:           f.Verified,
		AutosaveEnabled:    f.AutosaveEnabled,
		InstructionsFileID: f.InstructionsFileID,
		CreatedAt:          f.CreatedAt,
		UpdatedAt:          f.UpdatedAt,
	}
	if includeContent {
		resp.Content = f.Content
	}
	if f.InstructionsFile != nil {
		resp.Instructions = &instructionsSummary{
			ID:      f.InstructionsFile.ID,
			Name:    f.InstructionsFile.Name,
			Content: f.InstructionsFile.Content,
		}
	}
	return resp
}

func filesToListResponse(files []models.File) []fileResponse {
	out := make([]fileResponse, len(files))
	for i := range files {
		out[i] = fileToResponse(&files[i], true)
	}
	return out
}
