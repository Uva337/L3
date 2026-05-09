package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// ImageTask описывает задачу (сообщение), которую мы отправляем в Kafka
type ImageTask struct {
	ImageID string `json:"image_id"`
}

// KafkaProducer отвечает за отправку сообщений в брокер Kafka
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer создает и настраивает новый продюсер
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}

	return &KafkaProducer{writer: w}
}

// PublishImageTask отправляет ID картинки в очередь
func (p *KafkaProducer) PublishImageTask(ctx context.Context, imageID string) error {
	task := ImageTask{ImageID: imageID}

	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("ошибка сериализации задачи: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(imageID),
		Value: payload,
	}

	// Отправляем в Kafka!
	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения в Kafka: %w", err)
	}

	return nil
}

// Close аккуратно закрывает соединение при выключении сервера
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
