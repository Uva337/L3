package worker

import (
	"context"
	"fmt"
	"net/http"
	"notification-service/internal/storage"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationWorker struct {
	conn     *amqp.Connection
	repo     *storage.PostgresStorage
	botToken string
	chatID   string
}

// подключается к RabbitMQ
func NewNotificationWorker(url string, repo *storage.PostgresStorage, botToken string, chatID string) (*NotificationWorker, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения воркера к RabbitMQ: %w", err)
	}
	return &NotificationWorker{conn: conn, repo: repo, botToken: botToken, chatID: chatID}, nil
}

//запускает бесконечный цикл прослушивания очереди
func (w *NotificationWorker) Start() error {
	ch, err := w.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"notifications_queue", "", false, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	fmt.Println("⏳ Воркер запущен и ждет сообщения из RabbitMQ...")

	for d := range msgs {
		id := string(d.Body)
		fmt.Printf("\n📥 Воркер поймал задачу! ID: %s\n", id)

		notification, err := w.repo.GetByID(context.Background(), id)
		if err != nil {
			fmt.Printf("❌ Ошибка: не нашли уведомление в БД: %v\n", err)
			d.Nack(false, true)
			continue
		}

		fmt.Println("🚀 Отправляем сообщение в Telegram...")

		tgURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", w.botToken)
		tgBody := fmt.Sprintf(`{"chat_id": "%s", "text": "🔔 <b>Новая эсэмесочка!</b>\n\n<b>Текст:</b> %s", "parse_mode": "HTML"}`, w.chatID, notification.Message)

		resp, err := http.Post(tgURL, "application/json", strings.NewReader(tgBody))

		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Printf("❌ Ошибка отправки ТГ! Откладываем задачу %s (Exponential Backoff)...\n", id)

			w.repo.RescheduleNotification(context.Background(), id)

			d.Ack(false)

			if resp != nil {
				resp.Body.Close()
			}
			continue
		}
		resp.Body.Close()

		err = w.repo.UpdateStatus(context.Background(), id, "sent")
		if err != nil {
			fmt.Printf("❌ Ошибка БД: %v\n", err)
			d.Nack(false, true)
			continue
		}

		fmt.Printf("✅ Уведомление %s успешно отправлено в ТГ!\n", id)
		d.Ack(false)
	}

	return nil
}

func (w *NotificationWorker) Close() {
	if w.conn != nil {
		w.conn.Close()
	}
}
