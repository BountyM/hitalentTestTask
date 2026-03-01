package service

import "errors"

var (
	// ErrDepartmentNotFound - подразделение не найдено
	ErrDepartmentNotFound = errors.New("подразделение не найдено")
	// ErrInvalidParent - недопустимый родительский ID (например, попытка сделать подразделение своим собственным родителем)
	ErrInvalidParent = errors.New("недопустимый родительский ID")
	// ErrHierarchyCycle - обновление создаст циклическую ссылку в иерархии
	ErrHierarchyCycle = errors.New("обновление создаст циклическую ссылку в иерархии")
	// ErrInvalidReassignTarget - недопустимое подразделение для переназначения
	ErrInvalidReassignTarget = errors.New("недопустимое подразделение для переназначения")
	// ErrCannotReassignToChild - нельзя переназначить сотрудников в дочернее подразделение
	ErrCannotReassignToChild = errors.New("нельзя переназначить сотрудников в дочернее подразделение")
)
