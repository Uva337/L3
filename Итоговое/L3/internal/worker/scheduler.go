package worker

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/mq"
	"notification-service/internal/storage"
)

type Scheduler struct {
	repo     *storage.PostgresStorage
	producer *mq.Producer
}

// конструктор планировщика
func NewScheduler(repo *storage.PostgresStorage, producer *mq.Producer) *Scheduler {
	return &Scheduler{repo: repo, producer: producer}
}

// запускает бесконечный цикл проверок
func (s *Scheduler) Start() {
	fmt.Println("🕒 Планировщик запущен. Проверка каждую 10 секунд...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processDueNotifications()
		}
	}
}

func (s *Scheduler) processDueNotifications() {
	ctx := context.Background()
	notifications, err := s.repo.GetDueNotifications(ctx)
	if err != nil {
		fmt.Printf("❌ Ошибка поиска задач: %v\n", err)
		return
	}

	for _, n := range notifications {
		err := s.repo.UpdateStatus(ctx, n.ID, "queued")
		if err != nil {
			continue
		}

		err = s.producer.PublishID(ctx, n.ID)
		if err != nil {
			s.repo.UpdateStatus(ctx, n.ID, "pending")
			continue
		}

		fmt.Printf("⏰ Планировщик передал задачу %s в очередь брокера!\n", n.ID)
	}
}
