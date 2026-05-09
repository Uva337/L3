package worker

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"notification-service/internal/model"

	"github.com/nfnt/resize"
	"github.com/segmentio/kafka-go"
)

// Интерфейс для работы с БД из воркера
type ImageStorage interface {
	GetImage(ctx context.Context, id string) (*model.Image, error)
	UpdateImageStatus(ctx context.Context, id, status string) error
}

type ImageWorker struct {
	reader  *kafka.Reader
	storage ImageStorage
}

func NewImageWorker(brokers []string, topic, groupID string, storage ImageStorage) *ImageWorker {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &ImageWorker{
		reader:  r,
		storage: storage,
	}
}

// Start запускает бесконечный цикл чтения сообщений
func (w *ImageWorker) Start(ctx context.Context) {
	log.Println("Image Worker started, listening to Kafka...")

	for {
		msg, err := w.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("ошибка чтения из Kafka: %v\n", err)
			continue
		}

		var task struct {
			ImageID string `json:"image_id"`
		}
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			log.Printf("ошибка парсинга задачи: %v\n", err)
			continue
		}

		log.Printf("Воркер взял в работу картинку ID: %s\n", task.ImageID)

		err = w.processImage(ctx, task.ImageID)
		if err != nil {
			log.Printf("Ошибка обработки картинки %s: %v\n", task.ImageID, err)
			w.storage.UpdateImageStatus(ctx, task.ImageID, "failed")
		}
	}
}

// processImage выполняет физическую обработку файла
func (w *ImageWorker) processImage(ctx context.Context, imageID string) error {
	w.storage.UpdateImageStatus(ctx, imageID, "processing")

	imgInfo, err := w.storage.GetImage(ctx, imageID)
	if err != nil {
		return err
	}

	originalPath := filepath.Join("uploads", "originals", imgInfo.Filename)
	processedPath := filepath.Join("uploads", "processed", imgInfo.Filename)

	file, err := os.Open(originalPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// ШАГ 1: ДЕЛАЕМ РЕСАЙЗ (уменьшаем ширину до 800px)
	resizedImg := resize.Resize(800, 0, img, resize.Lanczos3)

	// ШАГ 2: НАКЛАДЫВАЕМ ВОДЯНОЙ ЗНАК
	watermarkedImg := addWatermark(resizedImg)

	// ШАГ 3: Сохраняем результат
	out, err := os.Create(processedPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Кодируем обратно (обрати внимание, теперь мы используем watermarkedImg)
	switch imgInfo.Format {
	case "jpg", "jpeg":
		err = jpeg.Encode(out, watermarkedImg, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(out, watermarkedImg)
	case "gif":
		err = gif.Encode(out, watermarkedImg, &gif.Options{})
	default:
		err = jpeg.Encode(out, watermarkedImg, nil)
	}

	if err != nil {
		return err
	}

	return w.storage.UpdateImageStatus(ctx, imageID, "ready")
}

func (w *ImageWorker) Close() error {
	return w.reader.Close()
}

func addWatermark(baseImg image.Image) image.Image {
	bounds := baseImg.Bounds()

	m := image.NewRGBA(bounds)

	draw.Draw(m, bounds, baseImg, image.Point{}, draw.Src)

	watermarkColor := color.RGBA{R: 0, G: 0, B: 0, A: 100}

	startX := bounds.Max.X - 150
	startY := bounds.Max.Y - 40

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	watermarkRect := image.Rect(startX, startY, bounds.Max.X, bounds.Max.Y)

	draw.Draw(m, watermarkRect, &image.Uniform{watermarkColor}, image.Point{}, draw.Over)

	return m
}
