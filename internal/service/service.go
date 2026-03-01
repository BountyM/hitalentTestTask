package service

import "github.com/BountyM/hitalentTestTask/internal/repository"

type Service struct {
	Department
}

func New(repository *repository.Repository) *Service {
	return &Service{
		Department: newDepartmentService(*repository),
	}
}
