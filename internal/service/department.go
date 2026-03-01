package service

import (
	"errors"
	"fmt"

	"github.com/BountyM/hitalentTestTask/internal/models"
	"github.com/BountyM/hitalentTestTask/internal/repository"
)

type DepartmentService struct {
	repository repository.Repository
}

func newDepartmentService(repository repository.Repository) *DepartmentService {
	return &DepartmentService{repository: repository}
}

//go:generate mockgen -source=department.go -destination=mocks/mock.go
type Department interface {
	CreateDepartment(department *models.Department) error
	Exists(departmentID int) (exists bool, err error)
	CreateEmployee(employee *models.Employee) error
	GetDepartment(departmentID int, depth int, includeEmployees bool) (*models.Department, error)
	UpdateDepartment(departmentID int, name *string, parentID *int) (*models.Department, error)
	DeleteDepartment(departmentID int, mode string, reassignToID *int) error
}

func (s *DepartmentService) CreateDepartment(department *models.Department) error {
	err := s.repository.CreateDepartment(department)
	if err != nil {
		return fmt.Errorf("DepartmentService CreateDepartment() %w", err)
	}
	return err
}

func (s *DepartmentService) Exists(departmentID int) (exists bool, err error) {
	exists, err = s.repository.Exists(departmentID)
	if err != nil {
		return false, fmt.Errorf("DepartmentService CreateDepartment() %w", err)
	}
	return
}

func (s *DepartmentService) CreateEmployee(employee *models.Employee) error {
	err := s.repository.CreateEmployee(employee)
	if err != nil {
		return fmt.Errorf("DepartmentService CreateEmployee() %w", err)
	}
	return err
}

// GetDepartment получает подразделение с полной иерархией
func (s *DepartmentService) GetDepartment(
	departmentID int,
	depth int,
	includeEmployees bool,
) (*models.Department, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 5 {
		depth = 5
	}

	deptartment, err := s.repository.GetDepartment(departmentID)
	if err != nil {
		return nil, ErrDepartmentNotFound
	}

	if includeEmployees {
		emps, err := s.repository.GetEmployeesByDepartmentID(departmentID)
		if err != nil {
			return nil, err
		}
		deptartment.Employees = emps
	}

	if depth > 1 {
		// Загружаем детей
		var children []models.Department
		if err := s.repository.GetDepartmentChildren(departmentID, &children); err != nil {
			return nil, err
		}

		for i := range children {
			childTree, err := s.GetDepartment(children[i].ID, depth-1, includeEmployees)
			if err != nil {
				continue
			}
			deptartment.Children = append(deptartment.Children, *childTree)
		}
	}
	return deptartment, nil
}

// UpdateDepartment обновляет подразделение (имя и/или родительское подразделение)
func (s *DepartmentService) UpdateDepartment(
	departmentID int,
	name *string,
	parentID *int,
) (*models.Department, error) {
	// Проверяем существование подразделения
	exists, err := s.repository.Exists(departmentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDepartmentNotFound
	}

	// Если parentID не указан или равен NULL, цикл невозможен
	if parentID == nil || *parentID == 0 {
		// Обновляем только имя, если оно есть
		return s.repository.UpdateDepartment(departmentID, name, parentID)
	}

	// Проверка 1: нельзя сделать подразделение родителем самого себя
	if *parentID == departmentID {
		return nil, ErrInvalidParent
	}

	// Проверка 2: проверяем существование родительского подразделения
	parentExists, err := s.repository.Exists(*parentID)
	if err != nil {
		return nil, err
	}
	if !parentExists {
		return nil, ErrInvalidParent
	}

	// Проверка 3: проверяем на циклические ссылки
	hasCycle, err := s.repository.HasCycle(departmentID, *parentID)
	if err != nil {
		return nil, err
	}
	if hasCycle {
		return nil, ErrHierarchyCycle
	}

	// Обновляем подразделение в репозитории
	updatedDept, err := s.repository.UpdateDepartment(departmentID, name, parentID)
	if err != nil {
		return nil, err
	}

	return updatedDept, nil
}

func (s *DepartmentService) DeleteDepartment(departmentID int, mode string, reassignToID *int) error {
	// Проверяем существование подразделения
	exists, err := s.repository.Exists(departmentID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDepartmentNotFound
	}

	switch mode {
	case "cascade":
		// Удаляем подразделение и всё его поддерево (включая сотрудников)
		return s.repository.DeleteDepartmentCascade(departmentID)

	case "reassign":
		if reassignToID == nil {
			return errors.New("reassign_to_department_id не указан")
		}

		// Проверяем, что подразделение для переназначения существует
		targetExists, err := s.repository.Exists(*reassignToID)
		if err != nil {
			return err
		}
		if !targetExists {
			return ErrInvalidReassignTarget
		}

		// Проверяем, что не пытаемся переназначить в дочернее подразделение (цикл)
		isChild, err := s.repository.IsChildDepartment(departmentID, *reassignToID)
		if err != nil {
			return err
		}
		if isChild {
			return ErrCannotReassignToChild
		}

		// Переназначаем сотрудников и удаляем подразделение
		return s.repository.DeleteDepartmentWithReassign(departmentID, *reassignToID)

	default:
		return errors.New("неизвестный режим удаления")
	}
}
