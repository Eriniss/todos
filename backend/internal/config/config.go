package config

import (
	"log"
	"os"
)

type Config struct {
	// Server
	ServerPort string

	// Database (PostgreSQL for on-premise, will migrate to DynamoDB)
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis Cache
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// RabbitMQ (will migrate to SQS)
	RabbitMQURL string

	// MongoDB
	MongoURL string
	MongoDB  string
}

func LoadConfig() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "3000"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "testbox"),
		DBPassword: getEnv("DB_PASSWORD", "testbox"),
		DBName:     getEnv("DB_NAME", "testbox"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,

		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),

		MongoURL: getEnv("MONGO_URL", "mongodb://localhost:27017"),
		MongoDB:  getEnv("MONGO_DB", "testbox"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("Environment variable %s not set, using default: %s", key, defaultValue)
	return defaultValue
}
