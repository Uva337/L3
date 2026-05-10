package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"notification-service/internal/model"

	"github.com/google/uuid"
)

type ImageStorage interface {
	SaveImage(ctx context.Context, img *model.Image) error
	GetImage(ctx context.Context, id string) (*model.Image, error)
	DeleteImage(ctx context.Context, id string) error
	GetAllImages(ctx context.Context) ([]*model.Image, error)
}

type ImageTaskProducer interface {
	PublishImageTask(ctx context.Context, imageID string) error
}

type ImageService struct {
	storage  ImageStorage
	producer ImageTaskProducer
}

func NewImageService(s ImageStorage, p ImageTaskProducer) *ImageService {
	return &ImageService{storage: s, producer: p}
}

func (s *ImageService) Upload(ctx context.Context, file *multipart.FileHeader) (*model.Image, error) {
	id := uuid.New().String()
	ext := strings.ToLower(filepath.Ext(file.Filename))
	format := strings.TrimPrefix(ext, ".")

	if format != "jpg" && format != "jpeg" && format != "png" && format != "gif" {
		return nil, fmt.Errorf("неподдерживаемый формат: %s (разрешены jpg, png, gif)", format)
	}
	if format == "jpeg" {
		format = "jpg"
	}

	filename := id + ext
	originalPath := filepath.Join("uploads", "originals", filename)

	// 2. Физически сохранение файла на диск
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(originalPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания файла на диске: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("ошибка записи файла: %w", err)
	}

	// 3. Сохраняем информацию в базу данных со статусом 'pending' (в ожидании)
	img := &model.Image{
		ID:        id,
		Filename:  filename,
		Status:    "pending",
		Format:    format,
		CreatedAt: time.Now(),
	}

	if err := s.storage.SaveImage(ctx, img); err != nil {
		os.Remove(originalPath)
		return nil, fmt.Errorf("ошибка сохранения в БД: %w", err)
	}

	// 4. Отправляем ID картинки в Kafka для фоновой обработки!
	if err := s.producer.PublishImageTask(ctx, id); err != nil {
		return nil, fmt.Errorf("ошибка отправки задачи в Kafka: %w", err)
	}

	return img, nil
}

// GetAll достает все картинки из базы для отображения на фронтенде
func (s *ImageService) GetAll(ctx context.Context) ([]*model.Image, error) {
	return s.storage.GetAllImages(ctx)
}

// Delete удаляет картинку из БД и чистит файлы с диска
func (s *ImageService) Delete(ctx context.Context, id string) error {
	img, err := s.storage.GetImage(ctx, id)
	if err != nil {
		return fmt.Errorf("картинка не найдена: %w", err)
	}

	if err := s.storage.DeleteImage(ctx, id); err != nil {
		return fmt.Errorf("ошибка удаления из БД: %w", err)
	}

	os.Remove(filepath.Join("uploads", "originals", img.Filename))
	os.Remove(filepath.Join("uploads", "processed", img.Filename))

	return nil
}

// GetByID возвращает информацию о картинке по её ID
func (s *ImageService) GetByID(ctx context.Context, id string) (*model.Image, error) {
	return s.storage.GetImage(ctx, id)
}
