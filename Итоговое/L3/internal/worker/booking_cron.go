package worker

import (
	"context"
	"log"
	"time"
)

type BookingCronStorage interface {
	CancelExpiredBookings(ctx context.Context) (int, error)
}

type BookingCron struct {
	storage  BookingCronStorage
	interval time.Duration
	ticker   *time.Ticker
	done     chan bool
}

// создает новый планировщик.
func NewBookingCron(storage BookingCronStorage, interval time.Duration) *BookingCron {
	return &BookingCron{
		storage:  storage,
		interval: interval,
		done:     make(chan bool),
	}
}

// запускает бесконечный цикл проверок
func (c *BookingCron) Start() {
	c.ticker = time.NewTicker(c.interval)
	log.Printf("Booking Cron started, checking for expired bookings every %v...", c.interval)

	for {
		select {
		case <-c.done:
			return
		case <-c.ticker.C:
			ctx := context.Background()
			count, err := c.storage.CancelExpiredBookings(ctx)
			if err != nil {
				log.Printf("[CRON ERROR] Ошибка при отмене броней: %v\n", err)
			} else if count > 0 {
				log.Printf("[CRON] Успешно отменено просроченных броней: %d. Места возвращены!\n", count)
			}
		}
	}
}

// останавливает крон при выключении сервера
func (c *BookingCron) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	c.done <- true
}
