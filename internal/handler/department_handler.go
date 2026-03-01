package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/BountyM/hitalentTestTask/internal/models"
	"github.com/BountyM/hitalentTestTask/internal/service"
)

type DepartmentHandler struct {
	services *service.Service
	logger   *slog.Logger
}

func NewDepartmentHandler(services *service.Service, logger *slog.Logger) *DepartmentHandler {
	return &DepartmentHandler{
		services: services,
		logger:   logger,
	}
}

func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {

	requestID := r.Context().Value(requestIDKey)

	h.logger.Info("Starting CreateDepartment handler",
		slog.Any("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	)

	var department models.Department
	if err := json.NewDecoder(r.Body).Decode(&department); err != nil {
		h.logger.Error("Failed to parse request body",
			slog.Any("request_id", requestID),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	// Валидация обязательного поля Name
	if strings.TrimSpace(department.Name) == "" {
		err := errors.New("name is required")
		h.logger.Error("Department name is required",
			slog.Any("request_id", requestID),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	if err := h.services.CreateDepartment(&department); err != nil {
		h.logger.Error("Failed to create department in service layer",
			slog.Any("request_id", requestID),
			slog.String("department_name", department.Name),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Логируем успешное создание подразделения
	h.logger.Info("Department created successfully",
		slog.Any("request_id", requestID),
		slog.Uint64("department_id", uint64(department.ID)),
		slog.String("department_name", department.Name),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(department); err != nil {
		h.logger.Error("Failed to encode department response",
			slog.Any("request_id", requestID),
			slog.String("error", err.Error()),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *DepartmentHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {

	requestID := r.Context().Value(requestIDKey)

	h.logger.Info("Starting CreateEmployee handler",
		slog.Any("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	)

	departmentIDStr := r.PathValue("id")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		h.logger.Error("Failed to extract department ID",
			slog.Any("request_id", requestID),
			slog.String("url", r.RequestURI),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, "Неверный ID подразделения", http.StatusBadRequest)
		return
	}

	// 2. Проверяем существование подразделения
	exists, err := h.services.Exists(departmentID)
	if err != nil {
		h.logger.Error("Failed to check department existence",
			slog.Any("request_id", requestID),
			slog.Uint64("department_id", uint64(departmentID)),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, "Ошибка проверки подразделения", http.StatusInternalServerError)
		return
	}
	if !exists {
		h.logger.Warn("Department not found",
			slog.Any("request_id", requestID),
			slog.Uint64("department_id", uint64(departmentID)),
		)
		writeJSONError(w, "Подразделение не найдено", http.StatusNotFound)
		return
	}

	// 3. Десериализуем тело запроса
	var employee models.Employee
	if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
		h.logger.Error("Failed to parse employee data",
			slog.Any("request_id", requestID),
			slog.Uint64("department_id", uint64(departmentID)),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, fmt.Sprintf("Ошибка парсинга JSON: %v", err), http.StatusBadRequest)
		return
	}

	// 4. Устанавливаем DepartmentID после десериализации
	employee.DepartmentID = departmentID

	// 5. Валидируем данные сотрудника
	if validationErr := validateEmployee(&employee); validationErr != nil {
		h.logger.Warn("Employee validation failed",
			slog.Any("request_id", requestID),
			slog.Any("employee", employee),
			slog.String("validation_error", validationErr.Error()),
		)
		writeJSONError(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	// 6. Создаём сотрудника
	if err := h.services.CreateEmployee(&employee); err != nil {
		h.logger.Error("Failed to create employee",
			slog.Any("request_id", requestID),
			slog.Any("employee", employee),
			slog.Uint64("department_id", uint64(departmentID)),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, "Ошибка создания сотрудника", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Employee created successfully",
		slog.Any("request_id", requestID),
		slog.Uint64("employee_id", uint64(employee.ID)),
		slog.String("employee_full_name", employee.FullName),
		slog.Uint64("department_id", uint64(departmentID)),
	)

	// 7. Возвращаем успешный ответ с данными сотрудника
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(employee); err != nil {
		h.logger.Error("Failed to encode department response",
			slog.Any("request_id", requestID),
			slog.String("error", err.Error()),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *DepartmentHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {

	requestID := r.Context().Value(requestIDKey)

	h.logger.Info("Starting GetDepartment handler",
		slog.Any("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	)

	// 1. Извлекаем ID подразделения из пути
	departmentIDStr := r.PathValue("id")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		h.logger.Error("Failed to extract department ID",
			slog.Any("request_id", requestID),
			slog.String("url", r.RequestURI),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, "Неверный ID подразделения", http.StatusBadRequest)
		return
	}

	// 2. Парсим query‑параметры
	depthStr := r.URL.Query().Get("depth")
	includeEmployeesStr := r.URL.Query().Get("include_employees")

	// Устанавливаем значения по умолчанию
	depth := 1
	includeEmployees := true

	if depthStr != "" {
		var err error
		depth, err = strconv.Atoi(depthStr)
		if err != nil || depth < 0 || depth > 5 {
			h.logger.Warn("Invalid depth parameter, using default",
				slog.Any("request_id", requestID),
				slog.String("provided_depth", depthStr),
				slog.Int("default_depth", 1),
			)
			depth = 1
		}
	}

	if includeEmployeesStr != "" {
		includeEmployees, err = strconv.ParseBool(includeEmployeesStr)
		if err != nil {
			h.logger.Warn("Invalid include_employees parameter, using default",
				slog.Any("request_id", requestID),
				slog.String("provided_include_employees", includeEmployeesStr),
				slog.Bool("default_include_employees", true),
			)
			includeEmployees = true
		}
	}

	department, err := h.services.GetDepartment(departmentID, depth, includeEmployees)
	if err != nil {
		if errors.Is(err, service.ErrDepartmentNotFound) {
			h.logger.Warn("Department not found",
				slog.Any("request_id", requestID),
				slog.Uint64("department_id", uint64(departmentID)),
			)
			writeJSONError(w, "Подразделение не найдено", http.StatusNotFound)
		} else {
			h.logger.Error("Failed to fetch department details",
				slog.Any("request_id", requestID),
				slog.Uint64("department_id", uint64(departmentID)),
				slog.String("error", err.Error()),
			)
			writeJSONError(w, "Ошибка получения данных подразделения", http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("Department details fetched successfully",
		slog.Any("request_id", requestID),
		slog.Uint64("department_id", uint64(department.ID)),
		slog.String("department_name", department.Name),
		slog.Int("employees_count", len(department.Employees)),
		slog.Int("children_count", len(department.Children)),
	)
	// 4. Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(department); err != nil {
		h.logger.Error("Failed to encode department response",
			slog.Any("request_id", requestID),
			slog.String("error", err.Error()),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(requestIDKey)

	h.logger.Info("Starting UpdateDepartment handler",
		slog.Any("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	)

	// 1. Извлекаем ID подразделения из пути
	departmentIDStr := r.PathValue("id")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		h.logger.Error("Failed to extract department ID",
			slog.Any("request_id", requestID),
			slog.String("url", r.RequestURI),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, "Неверный ID подразделения", http.StatusBadRequest)
		return
	}

	// 2. Парсим тело запроса
	var updateData struct {
		Name     *string `json:"name,omitempty"`
		ParentID *int    `json:"parent_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		h.logger.Error("Failed to parse request body",
			slog.Any("request_id", requestID),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, "Некорректный JSON в теле запроса", http.StatusBadRequest)
		return
	}

	// 3. Проверяем, что хотя бы одно поле для обновления присутствует
	if updateData.Name == nil && updateData.ParentID == nil {
		h.logger.Warn("No fields to update provided",
			slog.Any("request_id", requestID),
		)
		writeJSONError(w, "Необходимо указать хотя бы одно поле для обновления (name или parent_id)", http.StatusBadRequest)
		return
	}

	// 4. Вызываем сервис для обновления
	updatedDepartment, err := h.services.UpdateDepartment(departmentID, updateData.Name, updateData.ParentID)
	if err != nil {
		if errors.Is(err, service.ErrDepartmentNotFound) {
			h.logger.Warn("Department not found",
				slog.Any("request_id", requestID),
				slog.Int("department_id", departmentID),
			)
			writeJSONError(w, "Подразделение не найдено", http.StatusNotFound)
		} else if errors.Is(err, service.ErrInvalidParent) {
			h.logger.Warn("Invalid parent ID specified",
				slog.Any("request_id", requestID),
				slog.Int("department_id", departmentID),
				slog.Any("parent_id", updateData.ParentID),
			)
			writeJSONError(w, "Недопустимый родительский ID", http.StatusBadRequest)
		} else if errors.Is(err, service.ErrHierarchyCycle) {
			h.logger.Warn("Update create cycle",
				slog.Any("request_id", requestID),
				slog.Int("department_id", departmentID),
				slog.Any("parent_id", updateData.ParentID),
			)
			writeJSONError(w, "Обновление создаст циклическую ссылку в иерархии", http.StatusBadRequest)
		} else {
			h.logger.Error("Failed to update department",
				slog.Any("request_id", requestID),
				slog.Int("department_id", departmentID),
				slog.String("error", err.Error()),
			)
			writeJSONError(w, "Ошибка обновления подразделения", http.StatusInternalServerError)
		}
		return
	}

	// 5. Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	h.logger.Info("Department updated successfully",
		slog.Any("request_id", requestID),
		slog.Int("department_id", updatedDepartment.ID),
		slog.String("department_name", updatedDepartment.Name),
	)

	if err := json.NewEncoder(w).Encode(updatedDepartment); err != nil {
		h.logger.Error("Failed to encode response",
			slog.Any("request_id", requestID),
			slog.Int("department_id", updatedDepartment.ID),
			slog.String("error", err.Error()),
		)
		// Если ошибка произошла после установки статуса, ничего не можем сделать
		return
	}
}

func (h *DepartmentHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(requestIDKey)

	h.logger.Info("Starting DeleteDepartment handler",
		slog.Any("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	// Извлекаем ID подразделения из пути
	departmentIDStr := r.PathValue("id")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		h.logger.Error("Failed to extract department ID",
			slog.Any("request_id", requestID),
			slog.String("url", r.RequestURI),
			slog.String("error", err.Error()),
		)
		writeJSONError(w, "Неверный ID подразделения", http.StatusBadRequest)
		return
	}

	// Парсим query‑параметры
	mode := r.URL.Query().Get("mode")
	reassignToStr := r.URL.Query().Get("reassign_to_department_id")

	var reassignToID *int
	if reassignToStr != "" {
		reassignID, err := strconv.Atoi(reassignToStr)
		if err != nil {
			h.logger.Error("Invalid reassign_to_department_id",
				slog.Any("request_id", requestID),
				slog.String("value", reassignToStr),
				slog.String("error", err.Error()),
			)
			writeJSONError(w, "Недопустимый ID подразделения для переназначения", http.StatusBadRequest)
			return
		}
		reassignToID = &reassignID
	}

	// Валидация параметров
	if mode != "cascade" && mode != "reassign" {
		h.logger.Warn("Invalid delete mode",
			slog.Any("request_id", requestID),
			slog.String("mode", mode),
		)
		writeJSONError(w, "Режим удаления должен быть 'cascade' или 'reassign'", http.StatusBadRequest)
		return
	}

	if mode == "reassign" && reassignToID == nil {
		h.logger.Warn("Missing reassign_to_department_id for reassign mode",
			slog.Any("request_id", requestID),
		)
		writeJSONError(w, "Для режима 'reassign' необходимо указать reassign_to_department_id", http.StatusBadRequest)
		return
	}

	// Вызываем сервис для удаления
	err = h.services.DeleteDepartment(departmentID, mode, reassignToID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDepartmentNotFound):
			h.logger.Warn("Department not found",
				slog.Any("request_id", requestID),
				slog.Int("department_id", departmentID),
			)
			writeJSONError(w, "Подразделение не найдено", http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidReassignTarget):
			h.logger.Warn("Invalid reassign target",
				slog.Any("request_id", requestID),
				slog.Int("department_id", departmentID),
				slog.Any("reassign_to_id", reassignToID),
			)
			writeJSONError(w, "Недопустимое подразделение для переназначения сотрудников", http.StatusBadRequest)
		default:
			h.logger.Error("Failed to delete department",
				slog.Any("request_id", requestID),
				slog.Int("department_id", departmentID),
				slog.String("error", err.Error()),
			)
			writeJSONError(w, "Ошибка удаления подразделения", http.StatusInternalServerError)
		}
		return
	}

	// Успешное удаление — возвращаем 204 No Content
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("Department deleted successfully",
		slog.Any("request_id", requestID),
		slog.Int("department_id", departmentID),
		slog.String("mode", mode),
	)
}
