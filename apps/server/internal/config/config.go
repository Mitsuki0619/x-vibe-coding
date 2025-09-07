package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// サーバー設定
	Port string
	Env  string

	// データベース設定
	DatabaseURL     string
	TestDatabaseURL string

	// JWT設定
	JWTSecret string

	// CORS設定
	CORSOrigins []string

	// ログレベル
	LogLevel string
}

// Load は環境変数から設定を読み込みます
func Load() *Config {
	// .envファイルを読み込み（存在しない場合はスキップ）
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	config := &Config{
		Port:            getEnv("PORT", "8080"),
		Env:             getEnv("ENV", "development"),
		DatabaseURL:     getEnv("DATABASE_URL", "host=localhost user=sns_user password=sns_password dbname=sns_db port=5432 sslmode=disable"),
		TestDatabaseURL: getEnv("TEST_DATABASE_URL", "host=localhost user=sns_test_user password=sns_test_password dbname=sns_test_db port=5433 sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "default-secret-key-change-in-production"),
		CORSOrigins:     getCORSOrigins(),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}

	// 必須設定の検証
	if config.JWTSecret == "default-secret-key-change-in-production" && config.Env == "production" {
		log.Fatal("JWT_SECRET must be set in production environment")
	}

	return config
}

// LoadTest はテスト環境用の設定を読み込みます
func LoadTest() *Config {
	// .env.testファイルを読み込み（存在する場合）
	if err := godotenv.Load(".env.test"); err != nil {
		log.Printf("Warning: .env.test file not found, using default test config: %v", err)
	}

	config := Load() // 基本設定をロード

	// テスト環境用に上書き
	config.Env = "test"
	config.Port = getEnv("PORT", "8081")
	config.DatabaseURL = config.TestDatabaseURL
	config.LogLevel = getEnv("LOG_LEVEL", "debug")

	return config
}

// IsDevelopment は開発環境かどうかを判定します
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsTest はテスト環境かどうかを判定します
func (c *Config) IsTest() bool {
	return c.Env == "test"
}

// IsProduction は本番環境かどうかを判定します
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返します
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt は環境変数を整数として取得します
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool は環境変数をブール値として取得します
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getCORSOrigins はCORS設定を取得します
func getCORSOrigins() []string {
	origins := getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:19000")
	if origins == "" {
		return []string{"*"} // デフォルトは全て許可
	}

	return strings.Split(origins, ",")
}
