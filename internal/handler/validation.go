package handler

import (
	"errors"

	"github.com/BountyM/hitalentTestTask/internal/models"
)

func validateEmployee(employee *models.Employee) error {
	if employee.FullName == "" {
		return errors.New("поле full_name обязательно для заполнения")
	}
	if employee.Position == "" {
		return errors.New("поле position обязательно для заполнения")
	}
	return nil
}
