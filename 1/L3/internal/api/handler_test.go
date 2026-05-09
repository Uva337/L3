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

// Названия тестовых функций всегда должны начинаться со слова Test
func TestHandler_Create_InvalidJSON(t *testing.T) {
	// 1. Создаем наш хендлер.
	// Передаем nil вместо сервиса, так как при кривом JSON
	// код вернет ошибку ДО обращения к бизнес-логике и базе данных!
	h := NewHandler(nil)

	// 2. Формируем "кривой" JSON (нет обязательных полей recipient и scheduled_at)
	badJSON := []byte(`{"message": "Тест", "channel": "telegram"}`)

	// 3. Создаем виртуальный HTTP-запрос (имитируем Postman)
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(badJSON))
	req.Header.Set("Content-Type", "application/json")

	// 4. Создаем "записыватель" ответа (он поймает то, что ответит сервер)
	w := httptest.NewRecorder()

	// 5. Создаем роутер и регистрируем нашу ручку
	router := ginext.New("тестик")
	router.POST("/notify", h.Create)

	// 6. ВЫПОЛНЯЕМ ЗАПРОС!
	router.ServeHTTP(w, req)

	// 7. ПРОВЕРЯЕМ РЕЗУЛЬТАТ (Ассерт):
	// Мы ожидаем, что сервер не пропустит такой JSON и вернет статус 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус 400 (Bad Request), а получили %d", w.Code)
	}
}

// --- НАШ ФЕЙКОВЫЙ СЕРВИС (MOCK) ---
type mockService struct{}

func (m *mockService) CreateNotification(ctx context.Context, message, recipient, channel string, scheduledAt time.Time) (*model.Notification, error) {
	// Возвращаем фейковую структуру модели
	return &model.Notification{ID: "fake-id-777", Status: "pending"}, nil
}

func (m *mockService) GetStatus(ctx context.Context, id string) (*model.Notification, error) {
	// Возвращаем фейковую структуру модели
	return &model.Notification{ID: id, Status: "sent"}, nil
}

func (m *mockService) CancelNotification(ctx context.Context, id string) error {
	return nil
}

// --- ТЕСТ 1: Успешное создание (201 Created) ---
func TestHandler_Create_Success(t *testing.T) {
	// Подсовываем хендлеру нашу игрушку вместо базы данных!
	h := NewHandler(&mockService{})

	// Формируем ИДЕАЛЬНЫЙ JSON
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

// --- ТЕСТ 2: Успешное получение статуса (200 OK) ---
func TestHandler_GetStatus_Success(t *testing.T) {
	h := NewHandler(&mockService{})

	// Делаем GET запрос с выдуманным ID
	req := httptest.NewRequest(http.MethodGet, "/notify/fake-id-777", nil)
	w := httptest.NewRecorder()

	router := ginext.New("test-service")
	router.GET("/notify/:id", h.GetStatus)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200 (OK), а получили %d", w.Code)
	}
}
