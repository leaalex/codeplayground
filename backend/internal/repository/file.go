package repository

import (
	"goplayground/backend/internal/models"

	"gorm.io/gorm"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) List(userID uint) ([]models.File, error) {
	var files []models.File
	err := r.db.Where("user_id = ? AND name NOT LIKE ?", userID, "%.md").
		Order("updated_at DESC").Find(&files).Error
	return files, err
}

func (r *FileRepository) ListAll() ([]models.File, error) {
	var files []models.File
	err := r.db.Preload("User").Order("updated_at DESC").Find(&files).Error
	return files, err
}

func (r *FileRepository) GetByID(id uint) (*models.File, error) {
	var file models.File
	err := r.db.Preload("User").Preload("InstructionsFile").First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) DeleteByID(id uint) error {
	return r.db.Delete(&models.File{}, id).Error
}

func (r *FileRepository) Create(file *models.File) error {
	return r.db.Create(file).Error
}

func (r *FileRepository) Get(id uint, userID uint) (*models.File, error) {
	var file models.File
	err := r.db.Preload("InstructionsFile").
		Where("id = ? AND user_id = ?", id, userID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) Update(file *models.File) error {
	// Clear associations so Save does not overwrite FKs from preloaded relations
	// (e.g. changing UserID while User is still the previous owner).
	file.User = nil
	file.InstructionsFile = nil
	return r.db.Save(file).Error
}

func (r *FileRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.File{}).Error
}

func (r *FileRepository) CountInstructionsReferences(mdID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.File{}).Where("instructions_file_id = ?", mdID).Count(&count).Error
	return count, err
}

func (r *FileRepository) ClearInstructionsReferences(mdID uint) error {
	return r.db.Model(&models.File{}).Where("instructions_file_id = ?", mdID).
		Update("instructions_file_id", nil).Error
}
