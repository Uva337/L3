package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"notification-service/internal/api"
	"notification-service/internal/config"
	"notification-service/internal/mq"
	"notification-service/internal/service"
	"notification-service/internal/storage"
	"notification-service/internal/worker"

	cleanenvport "github.com/wb-go/wbf/config/cleanenv-port"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
	"github.com/wb-go/wbf/logger"
)

func main() {
	// 1. Загрузка конфигурации
	var cfg config.Config
	if err := cleanenvport.Load(&cfg); err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Инициализация логгер
	log, err := logger.InitLogger(
		logger.SlogEngine,
		"notification-service",
		cfg.Env,
		logger.WithLevel(logger.DebugLevel),
	)
	if err != nil {
		fmt.Printf("failed to init logger: %v\n", err)
		os.Exit(1)
	}

	// 3. Подключение к PostgreSQL
	db, err := pgxdriver.New(
		cfg.Postgres.DSN,
		log,
		pgxdriver.MaxPoolSize(10),
	)
	if err != nil {
		log.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Successfully connected to PostgreSQL!")

	// 4. Инициализация хранилища
	repo := storage.NewPostgresStorage(db)

	// 5. Подключение к RabbitMQ
	rabbitURL := "amqp://guest:guest@localhost:5672/"
	producer, err := mq.NewProducer(rabbitURL)
	if err != nil {
		log.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer producer.Close()
	log.Info("Successfully connected to RabbitMQ!")

	// 5.1 Инициализация и запуск Воркера
	notificationWorker, err := worker.NewNotificationWorker(rabbitURL, repo, cfg.Telegram.Token, cfg.Telegram.ChatID)
	if err != nil {
		log.Error("Failed to init worker", "error", err)
		os.Exit(1)
	}
	defer notificationWorker.Close()
	go notificationWorker.Start()

	// 5.2 Инициализация и запуск Планировщик
	notificationScheduler := worker.NewScheduler(repo, producer)
	go notificationScheduler.Start()

	// 5.3 Инициализация и запуск Redis
	redisStorage := storage.NewRedisStorage(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	log.Info("Successfully initialized Redis cache!")

	os.MkdirAll("uploads/originals", 0755)
	os.MkdirAll("uploads/processed", 0755)

	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "image-processing"

	// 5.4 Инициализация Kafka Продюсер
	kafkaProducer := mq.NewKafkaProducer(kafkaBrokers, kafkaTopic)
	defer kafkaProducer.Close()

	// 5.5 Инициализация и запуск Kafka Воркер
	imageWorker := worker.NewImageWorker(kafkaBrokers, kafkaTopic, "image-group-1", repo)
	defer imageWorker.Close()
	go imageWorker.Start(context.Background())

	// 6. Инициализация бизнес-логику
	notifierService := service.NewNotifierService(repo, producer, redisStorage)

	// 6.1 Инициализация бизнес-логику
	urlService := service.NewURLShortenerService(repo, redisStorage)

	// 6.2 Инициализация бизнес-логику (Сервис комментариев)
	commentService := service.NewCommentService(repo)

	// 6.3 (Сервис изображений)
	imageService := service.NewImageService(repo, kafkaProducer)

	bookingCron := worker.NewBookingCron(repo, 5*time.Second)
	go bookingCron.Start()
	defer bookingCron.Stop()

	// 7. Инициализация API (передаем сервисы)
	handler := api.NewHandler(notifierService)
	urlHandler := api.NewURLHandler(urlService)
	commentHandler := api.NewCommentHandler(commentService)

	imageHandler := api.NewImageHandler(imageService)
	eventService := service.NewEventService(repo)
	eventHandler := api.NewEventHandler(eventService)

	saleService := service.NewSaleService(repo)
	saleHandler := api.NewSaleHandler(saleService)

	warehouseService := service.NewWarehouseService(repo)
	warehouseHandler := api.NewWarehouseHandler(warehouseService)

	router := api.SetupRouter(handler, urlHandler, commentHandler, imageHandler, eventHandler, saleHandler, warehouseHandler, log)

	// 8. Запускаем HTTP-сервер
	log.Info("Starting HTTP server", "port", cfg.Server.Port)
	if err := router.Run(cfg.Server.Port); err != nil {
		log.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
