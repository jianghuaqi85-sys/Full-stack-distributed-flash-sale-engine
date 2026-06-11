package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	AppPort       int
	LogLevel      string
	AllowOrigins  []string
	Database      DatabaseConfig
	Redis         RedisConfig
	JWT           JWTConfig
	OpenTelemetry OpenTelemetryConfig
	Email         EmailConfig
	Kafka         KafkaConfig
	WsGateway     WsGatewayConfig
}

// WsGatewayConfig holds WebSocket Gateway configuration for distributed mode
type WsGatewayConfig struct {
	Enabled    bool   // Enable distributed WebSocket mode
	InstanceID string // Unique instance ID (auto-generated if empty)
}

type KafkaConfig struct {
	Brokers []string // Kafka broker 地址列表，如 ["localhost:9092"]
	Enabled bool    // 是否启用 Kafka（false 时降级为 Redis Stream）
}

type EmailConfig struct {
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string
	FromAddress string
	FromName    string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	ClusterMode  bool
	ClusterNodes []string // Redis Cluster 节点列表，如 ["node1:6379", "node2:6379"]
}

type JWTConfig struct {
	Secret string
	Expire int
}

type OpenTelemetryConfig struct {
	Endpoint     string
	ServiceName  string
	ExporterType string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:       getEnv("APP_ENV", "production"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		AllowOrigins: parseOrigins(getEnv("ALLOW_ORIGINS", "http://localhost:3000,http://localhost:5173")),
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default_secret_key_must_be_overridden_in_production"),
			Expire: getEnvInt("JWT_EXPIRE", 86400),
		},
		OpenTelemetry: OpenTelemetryConfig{
			Endpoint:     getEnv("OTEL_ENDPOINT", "localhost:4317"),
			ServiceName:  getEnv("OTEL_SERVICE_NAME", "order-system"),
			ExporterType: getEnv("OTEL_EXPORTER_TYPE", "stdout"),
		},
	}

	var err error
	cfg.AppPort, err = strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		cfg.AppPort = 8080
	}

	cfg.Database = DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		Name:            getEnv("DB_NAME", "order_system"),
		SSLMode:         getEnv("DB_SSLMODE", "require"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 50),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
		ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
		ConnMaxIdleTime: getEnvInt("DB_CONN_MAX_IDLE_TIME", 60),
	}
	cfg.Database.Port, err = strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		cfg.Database.Port = 5432
	}

	cfg.Redis = RedisConfig{
		Host:         getEnv("REDIS_HOST", "localhost"),
		Password:     getEnv("REDIS_PASSWORD", ""),
		ClusterMode:  getEnv("REDIS_CLUSTER_MODE", "false") == "true",
		ClusterNodes: parseBrokers(getEnv("REDIS_CLUSTER_NODES", "")),
	}
	cfg.Redis.Port, err = strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		cfg.Redis.Port = 6379
	}
	cfg.Redis.DB, err = strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		cfg.Redis.DB = 0
	}

	cfg.Email = EmailConfig{
		SMTPHost:    getEnv("SMTP_HOST", ""),
		Username:    getEnv("SMTP_USERNAME", ""),
		Password:    getEnv("SMTP_PASSWORD", ""),
		FromAddress: getEnv("SMTP_FROM_ADDRESS", ""),
		FromName:    getEnv("SMTP_FROM_NAME", "票务系统"),
	}
	cfg.Email.SMTPPort, err = strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		cfg.Email.SMTPPort = 587
	}

	cfg.Kafka = KafkaConfig{
		Brokers: parseBrokers(getEnv("KAFKA_BROKERS", "localhost:9092")),
		Enabled: getEnv("KAFKA_ENABLED", "false") == "true",
	}

	cfg.WsGateway = WsGatewayConfig{
		Enabled:    getEnv("WS_GATEWAY_ENABLED", "false") == "true",
		InstanceID: getEnv("WS_INSTANCE_ID", ""),
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.AppEnv == "production" {
		if c.JWT.Secret == "default_secret_key_must_be_overridden_in_production" || c.JWT.Secret == "" {
			return errors.New("JWT_SECRET must be set in production environment")
		}
	}

	if c.Database.MaxOpenConns <= 0 {
		return errors.New("DB_MAX_OPEN_CONNS must be greater than 0")
	}

	if c.Database.MaxIdleConns < 0 {
		return errors.New("DB_MAX_IDLE_CONNS must be non-negative")
	}

	if c.Database.ConnMaxLifetime <= 0 {
		return errors.New("DB_CONN_MAX_LIFETIME must be greater than 0")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func parseOrigins(s string) []string {
	parts := strings.Split(s, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

func parseBrokers(s string) []string {
	parts := strings.Split(s, ",")
	brokers := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			brokers = append(brokers, p)
		}
	}
	return brokers
}
