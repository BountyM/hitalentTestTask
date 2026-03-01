package repository

import "gorm.io/gorm"

type Repository struct {
	Department
}

func New(db *gorm.DB) *Repository {
	return &Repository{
		Department: NewDepartmentPostgres(db),
	}
}
