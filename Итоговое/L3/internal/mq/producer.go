package mq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewProducer(url string) (*Producer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия канала: %w", err)
	}

	// создает очередь, если её еще нет
	_, err = ch.QueueDeclare(
		"notifications_queue", // Имя очереди
		true,                  // durable (очередь переживет перезагрузку RabbitMQ)
		false,                 // auto-delete (не удалять, если нет слушателей)
		false,                 // exclusive (не блокировать для других подключений)
		false,                 // no-wait
		nil,                   // аргументы
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка объявления очереди: %w", err)
	}

	return &Producer{conn: conn, channel: ch}, nil
}

// PublishID отправляет ID уведомления в очередь
func (p *Producer) PublishID(ctx context.Context, id string) error {
	err := p.channel.PublishWithContext(ctx,
		"",                    // exchange (используем дефолтный)
		"notifications_queue", // routing key (имя нашей очереди)
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:         []byte(id),
		})
	return err
}

func (p *Producer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
