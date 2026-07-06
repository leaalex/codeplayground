package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name      string         `gorm:"not null" json:"name"`
	Path      string         `gorm:"default:''" json:"path"`
	Content         string         `gorm:"type:text" json:"content"`
	Verified        bool           `gorm:"default:false" json:"verified"`
	AutosaveEnabled bool           `gorm:"default:true" json:"autosave_enabled"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (File) TableName() string {
	return "files"
}
