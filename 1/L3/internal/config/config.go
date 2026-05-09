package config

type Config struct {
	Env    string `yaml:"env" env-default:"local"`
	Server struct {
		Port string `yaml:"port" env-default:":8080"`
	} `yaml:"server"`
	Postgres struct {
		DSN string `yaml:"dsn" env-required:"true"`
	} `yaml:"postgres"`
	RabbitMQ struct {
		URL string `yaml:"url" env-required:"true"`
	} `yaml:"rabbitmq"`
	Telegram TelegramConfig `yaml:"telegram"`
	Redis    RedisConfig    `yaml:"redis"`
}

type TelegramConfig struct {
	Token  string `yaml:"token"`
	ChatID string `yaml:"chat_id"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
