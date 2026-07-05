package model

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// mysqlDuplicateEntryErrorCode is the MySQL error number for a unique/primary key violation.
const mysqlDuplicateEntryErrorCode = 1062

// IsDuplicateEntry reports whether err is a MySQL duplicate-key violation.
func IsDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError

	return errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntryErrorCode
}

type Repository interface {
	Create(email *Email) error
	Update(email *Email) error
	UpdateStatus(id uint, status EmailStatus) error
	FindByTo(to string) (*Email, error)
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

func (repository *repository) FindByTo(to string) (*Email, error) {
	var email Email
	result := repository.db.Where("`to` = ?", to).First(&email)
	return &email, result.Error
}
