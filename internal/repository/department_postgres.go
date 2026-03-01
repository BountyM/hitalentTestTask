package repository

import (
	"errors"
	"fmt"

	"github.com/BountyM/hitalentTestTask/internal/models"

	"gorm.io/gorm"
)

type Department interface {
	CreateDepartment(department *models.Department) error
	Exists(departmentID int) (exists bool, err error)
	CreateEmployee(employee *models.Employee) error

	GetDepartment(departmentID int) (*models.Department, error)
	GetEmployeesByDepartmentID(departmentID int) ([]models.Employee, error)
	GetDepartmentChildren(departmentID int, children *[]models.Department) error

	UpdateDepartment(departmentID int, name *string, parentID *int) (*models.Department, error)
	HasCycle(departmentID int, newParentID int) (bool, error)
	DeleteDepartmentCascade(departmentID int) error
	DeleteDepartmentWithReassign(departmentID int, reassignToID int) error
	IsChildDepartment(sourceDepartmentID int, targetDepartmentID int) (bool, error)
}

type DepartmentPostgres struct {
	db *gorm.DB
}

func NewDepartmentPostgres(db *gorm.DB) *DepartmentPostgres {
	return &DepartmentPostgres{
		db: db,
	}
}

func (r *DepartmentPostgres) CreateDepartment(department *models.Department) error {
	result := r.db.Create(department)
	if result.Error != nil {
		return fmt.Errorf("DepartmentPostgres CreateDepartment() ошибка: %w", result.Error)
	}
	return nil
}

func (r *DepartmentPostgres) Exists(departmentID int) (exists bool, err error) {

	err = r.db.Raw(
		"SELECT EXISTS(SELECT 1 FROM departments WHERE id = ?)",
		departmentID,
	).Scan(&exists).Error

	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования подразделения: %w", err)
	}

	return exists, nil
}

func (r *DepartmentPostgres) CreateEmployee(employee *models.Employee) error {
	result := r.db.Create(employee)
	if result.Error != nil {
		return fmt.Errorf("DepartmentPostgres CreateDepartment() ошибка: %w", result.Error)
	}
	return nil
}

func (r *DepartmentPostgres) GetDepartment(
	departmentID int,
) (*models.Department, error) {
	var department models.Department

	// Базовый запрос с предварительной загрузкой связей:
	// - дочерние подразделения (Children)
	// - сотрудники подразделения (Employees), отсортированные по created_at и full_name
	err := r.db.Preload("Children").
		Preload("Employees", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC, full_name ASC")
		}).
		First(&department, departmentID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDepartmentNotFound
		}
		return nil, err
	}

	return &department, nil
}

var ErrDepartmentNotFound = errors.New("подразделение не найдено")

func (r *DepartmentPostgres) GetEmployeesByDepartmentID(
	departmentID int,
) ([]models.Employee, error) {
	var employees []models.Employee

	// Запрос на получение всех сотрудников подразделения
	// с сортировкой по дате создания и ФИО
	err := r.db.Where("department_id = ?", departmentID).
		Order("created_at ASC, full_name ASC").
		Find(&employees).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если сотрудников нет, возвращаем пустой срез
			return []models.Employee{}, nil
		}
		return nil, err
	}

	return employees, nil
}

func (r *DepartmentPostgres) GetDepartmentChildren(
	departmentID int,
	children *[]models.Department,
) error {
	// Очистка целевого среза перед заполнением
	*children = []models.Department{}

	// Запрос на получение всех дочерних подразделений заданного подразделения
	// с предварительной загрузкой сотрудников каждого дочернего подразделения
	err := r.db.Where("parent_id = ?", departmentID).
		Preload("Employees", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC, full_name ASC")
		}).
		Find(children).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если дочерних подразделений нет, оставляем пустой срез
			*children = []models.Department{}
			return nil
		}
		return err
	}

	return nil
}

// UpdateDepartment обновляет подразделение в БД
func (r *DepartmentPostgres) UpdateDepartment(
	departmentID int,
	name *string,
	parentID *int,
) (*models.Department, error) {
	var department models.Department

	updates := make(map[string]interface{})

	if name != nil {
		updates["name"] = *name
	}
	if parentID != nil {
		updates["parent_id"] = *parentID
	}

	// Проверяем, есть ли что обновлять
	if len(updates) == 0 {
		return nil, errors.New("нет данных для обновления")
	}

	// Выполняем обновление
	result := r.db.Model(&department).
		Where("id = ?", departmentID).
		Updates(updates)

	if result.Error != nil {
		return nil, result.Error
	}

	// Если ничего не обновилось (например, подразделение не найдено)
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Загружаем обновлённую запись с связями
	err := r.db.Preload("Children").
		Preload("Employees").
		First(&department, departmentID).Error

	if err != nil {
		return nil, err
	}

	return &department, nil
}

// hasCycle проверяет, не создаст ли изменение parent_id циклическую ссылку
func (r *DepartmentPostgres) HasCycle(departmentID int, newParentID int) (bool, error) {
	currentID := newParentID

	for currentID != 0 {
		var parent models.Department
		err := r.db.Select("id", "parent_id").
			First(&parent, currentID).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil // если родитель не найден, цикла нет
			}
			return false, err
		}

		// Если нашли исходное подразделение в цепочке предков - цикл!
		if parent.ID == departmentID {
			return true, nil
		}

		// Переходим к следующему родителю
		if parent.ParentID == nil {
			break
		}
		currentID = *parent.ParentID
	}

	return false, nil
}

// DeleteDepartmentCascade удаляет подразделение, всех сотрудников и все дочерние подразделения (каскадно)
func (r *DepartmentPostgres) DeleteDepartmentCascade(departmentID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Удаляем всех сотрудников подразделения и его дочерних подразделений
		if err := tx.Exec(`DELETE FROM employees WHERE department_id IN (
		WITH RECURSIVE subtree AS (
		SELECT id FROM departments WHERE id = ?
		UNION ALL
		SELECT d.id FROM departments d
		JOIN subtree s ON d.parent_id = s.id
	) SELECT id FROM subtree)`, departmentID).Error; err != nil {
			return err
		}

		// 2. Удаляем все дочерние подразделения рекурсивно
		if err := tx.Exec(`WITH RECURSIVE to_delete AS (
	SELECT id FROM departments WHERE id = ?
	UNION ALL
	SELECT d.id FROM departments d
	JOIN to_delete td ON d.parent_id = td.id
	) DELETE FROM departments WHERE id IN (SELECT id FROM to_delete)`, departmentID).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteDepartmentWithReassign удаляет подразделение, но переназначает сотрудников в другое подразделение
func (r *DepartmentPostgres) DeleteDepartmentWithReassign(departmentID int, reassignToID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Переназначаем сотрудников подразделения на новое подразделение
		if err := tx.Model(&models.Employee{}).
			Where("department_id = ?", departmentID).
			Update("department_id", reassignToID).Error; err != nil {
			return err
		}

		// 2. Обновляем parent_id у дочерних подразделений, чтобы они стали дочерними для нового подразделения
		if err := tx.Model(&models.Department{}).
			Where("parent_id = ?", departmentID).
			Update("parent_id", reassignToID).Error; err != nil {
			return err
		}

		// 3. Удаляем само подразделение
		if err := tx.Delete(&models.Department{}, departmentID).Error; err != nil {
			return err
		}

		return nil
	})
}

// IsChildDepartment проверяет, является ли targetDepartment дочерним подразделением sourceDepartment
func (r *DepartmentPostgres) IsChildDepartment(sourceDepartmentID int, targetDepartmentID int) (bool, error) {
	var exists bool
	err := r.db.Raw(`
	WITH RECURSIVE subtree AS (
		SELECT id, parent_id FROM departments WHERE id = ?
	UNION ALL
		SELECT d.id, d.parent_id FROM departments d
		JOIN subtree s ON d.parent_id = s.id
	)
	SELECT EXISTS(SELECT 1 FROM subtree WHERE id = ?)
	`, sourceDepartmentID, targetDepartmentID).Scan(&exists).Error

	if err != nil {
		return false, err
	}
	return exists, nil
}
