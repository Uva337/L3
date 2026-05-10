package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type ImageTask struct {
	ImageID string `json:"image_id"`
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}

	return &KafkaProducer{writer: w}
}

// отправляет ID картинки в очередь
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

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения в Kafka: %w", err)
	}

	return nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
