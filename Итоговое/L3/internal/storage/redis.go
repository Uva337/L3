package storage

import (
	"context"
	"encoding/json"
	"time"

	"notification-service/internal/model" 

	"github.com/redis/go-redis/v9"
)

// обертка над клиентом Redis
type RedisStorage struct {
	client *redis.Client
}

//  конструктор подключения
func NewRedisStorage(addr, password string, db int) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisStorage{client: client}
}

//охраняем уведомление в кэш на 1 час
func (r *RedisStorage) SetNotification(ctx context.Context, notif *model.Notification) error {

	data, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	// Ключ будет выглядеть как "notify:123"
	return r.client.Set(ctx, "notify:"+notif.ID, data, time.Hour).Err()
}

//  пытаемся найти уведомление в кэше
func (r *RedisStorage) GetNotification(ctx context.Context, id string) (*model.Notification, error) {
	val, err := r.client.Get(ctx, "notify:"+id).Result()
	if err != nil {
		return nil, err
	}

	var notif model.Notification
	err = json.Unmarshal([]byte(val), &notif)
	return &notif, err
}

func (r *RedisStorage) SetURL(ctx context.Context, shortCode, originalURL string) error {
	return r.client.Set(ctx, "url:"+shortCode, originalURL, 24*time.Hour).Err()
}

//  получаем длинный URL из кэша
func (r *RedisStorage) GetURL(ctx context.Context, shortCode string) (string, error) {
	return r.client.Get(ctx, "url:"+shortCode).Result()
}
