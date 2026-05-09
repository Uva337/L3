package main

import (
	"fmt"
	"os"

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
	// 1. Загружаем конфигурацию
	var cfg config.Config
	if err := cleanenvport.Load(&cfg); err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Инициализируем логгер
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

	// 3. Подключаемся к PostgreSQL
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

	// 4. Инициализируем хранилище
	repo := storage.NewPostgresStorage(db)

	// 5. Подключаемся к RabbitMQ
	rabbitURL := "amqp://guest:guest@localhost:5672/"
	producer, err := mq.NewProducer(rabbitURL)
	if err != nil {
		log.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer producer.Close()
	log.Info("Successfully connected to RabbitMQ!")

	// 5.5 Инициализируем и запускаем Воркер
	notificationWorker, err := worker.NewNotificationWorker(rabbitURL, repo, cfg.Telegram.Token, cfg.Telegram.ChatID)
	if err != nil {
		log.Error("Failed to init worker", "error", err)
		os.Exit(1)
	}
	defer notificationWorker.Close()
	go notificationWorker.Start()

	// 5.6 Инициализируем и запускаем Планировщик
	notificationScheduler := worker.NewScheduler(repo, producer)
	go notificationScheduler.Start()

	// 5.7 Инициализируем хранилище Redis (Наш сверхбыстрый кэш)
	redisStorage := storage.NewRedisStorage(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	log.Info("Successfully initialized Redis cache!")

	// 6. Инициализируем бизнес-логику (передаем БД, Брокера И КЭШ REDIS)
	notifierService := service.NewNotifierService(repo, producer, redisStorage)

	// 7. Инициализируем API (передаем сервис)
	handler := api.NewHandler(notifierService)
	router := api.SetupRouter(handler, log)

	// 8. Запускаем HTTP-сервер
	log.Info("Starting HTTP server", "port", cfg.Server.Port)
	if err := router.Run(cfg.Server.Port); err != nil {
		log.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
