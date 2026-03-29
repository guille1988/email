package model

import (
	"gorm.io/gorm"
)

type Repository interface {
	Create(email *Email) error
	Update(email *Email) error
	UpdateStatus(id uint, status EmailStatus) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (repository *repository) Create(email *Email) error {
	return repository.db.Create(email).Error
}

func (repository *repository) Update(email *Email) error {
	return repository.db.Save(email).Error
}

func (repository *repository) UpdateStatus(id uint, status EmailStatus) error {
	return repository.db.Model(&Email{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}
