package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"notification-service/internal/model"
	"testing"
	"time"

	"github.com/wb-go/wbf/ginext"
)

func TestHandler_Create_InvalidJSON(t *testing.T) {
	h := NewHandler(nil)

	badJSON := []byte(`{"message": "Тест", "channel": "telegram"}`)

	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(badJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router := ginext.New("тестик")
	router.POST("/notify", h.Create)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус 400 (Bad Request), а получили %d", w.Code)
	}
}

type mockService struct{}

func (m *mockService) CreateNotification(ctx context.Context, message, recipient, channel string, scheduledAt time.Time) (*model.Notification, error) {
	return &model.Notification{ID: "fake-id-777", Status: "pending"}, nil
}

func (m *mockService) GetStatus(ctx context.Context, id string) (*model.Notification, error) {
	return &model.Notification{ID: id, Status: "sent"}, nil
}

func (m *mockService) CancelNotification(ctx context.Context, id string) error {
	return nil
}

// ТЕСТ 1: Успешное создание (201 Created) 
func TestHandler_Create_Success(t *testing.T) {
	h := NewHandler(&mockService{})

	validJSON := []byte(`{"message": "Всё работает!", "recipient": "user@test.com", "channel": "telegram", "scheduled_at": "2026-12-31T23:59:59Z"}`)

	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(validJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := ginext.New("test-service")
	router.POST("/notify", h.Create)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Ожидался статус 201 (Created), а получили %d", w.Code)
	}
}

// ТЕСТ 2: Успешное получение статуса (200 OK) 
func TestHandler_GetStatus_Success(t *testing.T) {
	h := NewHandler(&mockService{})

	req := httptest.NewRequest(http.MethodGet, "/notify/fake-id-777", nil)
	w := httptest.NewRecorder()

	router := ginext.New("test-service")
	router.GET("/notify/:id", h.GetStatus)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200 (OK), а получили %d", w.Code)
	}
}
