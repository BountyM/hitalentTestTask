package handler_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BountyM/hitalentTestTask/internal/handler"
	"github.com/BountyM/hitalentTestTask/internal/models"
	"github.com/BountyM/hitalentTestTask/internal/service"
	mock_service "github.com/BountyM/hitalentTestTask/internal/service/mocks"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/golang/mock/gomock"
)

func TestDepartmentHandler_CreateDepartment(t *testing.T) {
	type mockBehavior func(s *mock_service.MockDepartment)

	tests := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
		expectedResponse   string
		mockBehavior       mockBehavior
	}{
		{
			name:               "OK",
			requestBody:        `{"name":"Backend"}`,
			expectedStatusCode: http.StatusCreated, // 201
			expectedResponse:   `{"id":1,"name":"Backend","created_at":"0001-01-01T00:00:00Z","employees":null}`,
			mockBehavior: func(s *mock_service.MockDepartment) {
				s.EXPECT().
					CreateDepartment(gomock.Any()).
					DoAndReturn(func(d *models.Department) error {
						// Проверяем, что имя передано правильно
						if d.Name != "Backend" {
							return fmt.Errorf("expected name 'Backend', got '%s'", d.Name)
						}
						// Имитируем присвоение ID базой данных
						d.ID = 1
						return nil
					})
			},
		},
		{
			name:               "Missing name",
			requestBody:        `{"name":""}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: `{
									"error": "Error: name is required",
									"status": "Bad Request"
								}`,
			mockBehavior: nil, // вызов сервиса не ожидается
		},
		{
			name:               "Invalid JSON",
			requestBody:        `{"name":`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: `{
									"error": "Error: unexpected EOF",
									"status": "Bad Request"
								}`,
			mockBehavior: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockDepartment(ctrl)
			if tt.mockBehavior != nil {
				tt.mockBehavior(mockService)
			}

			services := &service.Service{Department: mockService}
			handler := handler.NewDepartmentHandler(services, slog.Default())

			req := httptest.NewRequest(http.MethodPost, "/departments", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateDepartment(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.JSONEq(t, tt.expectedResponse, w.Body.String())
		})
	}
}
