package service

import (
	"context"
	"math/rand"
	"time"

	"notification-service/internal/model"

	"github.com/google/uuid"
)

// Набор символов для коротких ссылок
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateShortCode генерирует случайную строку заданной длины
func generateShortCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// интерфейс для работы с PostgreSQL
type URLStorage interface {
	SaveURL(ctx context.Context, u *model.URL) error
	GetURLByCode(ctx context.Context, code string) (*model.URL, error)
	SaveClick(ctx context.Context, click *model.Click) error
	GetClicksCount(ctx context.Context, urlID string) (int, error)
	GetClickStats(ctx context.Context, urlID string) (map[string]int, error)
}

// нтерфейс для работы с Redis
type URLCache interface {
	SetURL(ctx context.Context, shortCode, originalURL string) error
	GetURL(ctx context.Context, shortCode string) (string, error)
}

type URLShortenerService struct {
	storage URLStorage
	cache   URLCache
}

func NewURLShortenerService(s URLStorage, c URLCache) *URLShortenerService {
	// Инициализируем генератор чисел
	rand.Seed(time.Now().UnixNano())
	return &URLShortenerService{
		storage: s,
		cache:   c,
	}
}

// создает новую короткую ссылку
func (s *URLShortenerService) ShortenURL(ctx context.Context, originalURL, customAlias string) (*model.URL, error) {
	shortCode := customAlias
	if shortCode == "" {
		shortCode = generateShortCode(6)
	}

	url := &model.URL{
		ID:          uuid.New().String(),
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	err := s.storage.SaveURL(ctx, url)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetURL(ctx, shortCode, originalURL)

	return url, nil
}

// ProcessRedirect теперь работает с Redis и PostgreSQL
func (s *URLShortenerService) ProcessRedirect(ctx context.Context, shortCode, userAgent string) (string, error) {
	// 1. Сначала проверяем Redis (Cache Hit)
	cachedURL, err := s.cache.GetURL(ctx, shortCode)
	if err == nil && cachedURL != "" {

		go s.asyncRecordClickByCode(shortCode, userAgent)
		return cachedURL, nil
	}

	// 2. Если в кэше нет — идем в базу
	url, err := s.storage.GetURLByCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	// 3. Сохраняем в кэш для будущих запросов
	_ = s.cache.SetURL(ctx, shortCode, url.OriginalURL)

	// 4. Записываем аналитику асинхронно
	s.recordClick(url.ID, userAgent)

	return url.OriginalURL, nil
}

// возвращает статистику по ссылке
func (s *URLShortenerService) GetAnalytics(ctx context.Context, shortCode string) (map[string]interface{}, error) {
	url, err := s.storage.GetURLByCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	clicks, err := s.storage.GetClicksCount(ctx, url.ID)
	if err != nil {
		return nil, err
	}

	uaStats, err := s.storage.GetClickStats(ctx, url.ID)
	if err != nil {
		uaStats = make(map[string]int)
	}

	return map[string]interface{}{
		"short_code":   url.ShortCode,
		"original_url": url.OriginalURL,
		"total_clicks": clicks,
		"user_agents":  uaStats,
		"created_at":   url.CreatedAt,
	}, nil
}

// асинхронная запись клика по ID
func (s *URLShortenerService) recordClick(urlID, userAgent string) {
	go func() {
		click := &model.Click{
			ID:        uuid.New().String(),
			URLID:     urlID,
			UserAgent: userAgent,
			ClickedAt: time.Now(),
		}
		_ = s.storage.SaveClick(context.Background(), click)
	}()
}

// когда нашли в кэше, нам всё равно нужен ID для аналитики
func (s *URLShortenerService) asyncRecordClickByCode(code, userAgent string) {
	url, err := s.storage.GetURLByCode(context.Background(), code)
	if err == nil {
		s.recordClick(url.ID, userAgent)
	}
}
