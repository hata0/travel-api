package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config はアプリケーション全体の設定を表すインターフェース
type Config interface {
	Database() DatabaseConfig
	JWT() JWTConfig
	Server() ServerConfig
	Log() LogConfig
	Environment() string
	Version() string
	IsProduction() bool
	IsDevelopment() bool
	Validate() error
}

// appConfig は Config インターフェースの実装
type appConfig struct {
	database    DatabaseConfig
	jwt         JWTConfig
	server      ServerConfig
	log         LogConfig
	environment string
	version     string
}

// DatabaseConfig はデータベース設定
type DatabaseConfig interface {
	URL() string
	MaxConnections() int
	MaxIdleTime() time.Duration
	ConnMaxLifetime() time.Duration
}

// JWTConfig はJWT設定
type JWTConfig interface {
	Secret() string
	AccessTokenExpiration() time.Duration
	RefreshTokenExpiration() time.Duration
	Issuer() string
}

// ServerConfig はサーバー設定
type ServerConfig interface {
	Port() string
	Host() string
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
	IdleTimeout() time.Duration
	ShutdownTimeout() time.Duration
	Address() string
}

// LogConfig はログ設定
type LogConfig interface {
	Level() slog.Level
	Format() string
	AddSource() bool
}

// 具体的な実装
type databaseConfig struct {
	url             string
	maxConnections  int
	maxIdleTime     time.Duration
	connMaxLifetime time.Duration
}

func (d databaseConfig) URL() string                    { return d.url }
func (d databaseConfig) MaxConnections() int            { return d.maxConnections }
func (d databaseConfig) MaxIdleTime() time.Duration     { return d.maxIdleTime }
func (d databaseConfig) ConnMaxLifetime() time.Duration { return d.connMaxLifetime }

type jwtConfig struct {
	secret                 string
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
	issuer                 string
}

func (j jwtConfig) Secret() string                        { return j.secret }
func (j jwtConfig) AccessTokenExpiration() time.Duration  { return j.accessTokenExpiration }
func (j jwtConfig) RefreshTokenExpiration() time.Duration { return j.refreshTokenExpiration }
func (j jwtConfig) Issuer() string                        { return j.issuer }

type serverConfig struct {
	port            string
	host            string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
}

func (s serverConfig) Port() string                   { return s.port }
func (s serverConfig) Host() string                   { return s.host }
func (s serverConfig) ReadTimeout() time.Duration     { return s.readTimeout }
func (s serverConfig) WriteTimeout() time.Duration    { return s.writeTimeout }
func (s serverConfig) IdleTimeout() time.Duration     { return s.idleTimeout }
func (s serverConfig) ShutdownTimeout() time.Duration { return s.shutdownTimeout }
func (s serverConfig) Address() string                { return s.host + ":" + s.port }

type logConfig struct {
	level     slog.Level
	format    string
	addSource bool
}

func (l logConfig) Level() slog.Level { return l.level }
func (l logConfig) Format() string    { return l.format }
func (l logConfig) AddSource() bool   { return l.addSource }

// appConfig のメソッド実装
func (c appConfig) Database() DatabaseConfig { return c.database }
func (c appConfig) JWT() JWTConfig           { return c.jwt }
func (c appConfig) Server() ServerConfig     { return c.server }
func (c appConfig) Log() LogConfig           { return c.log }
func (c appConfig) Environment() string      { return c.environment }
func (c appConfig) Version() string          { return c.version }
func (c appConfig) IsProduction() bool       { return c.environment == "production" }
func (c appConfig) IsDevelopment() bool      { return c.environment == "development" }

// ConfigError は設定エラーを表す
type ConfigError struct {
	Field   string
	Value   string
	Message string
}

func (e *ConfigError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("config validation failed [%s=%s]: %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("config validation failed [%s]: %s", e.Field, e.Message)
}

// ValidationErrors は複数のバリデーションエラーを格納する
type ValidationErrors struct {
	Errors []*ConfigError
}

func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "no validation errors"
	}
	if len(v.Errors) == 1 {
		return v.Errors[0].Error()
	}

	var builder strings.Builder
	builder.WriteString("multiple validation errors:\n")
	for _, err := range v.Errors {
		builder.WriteString("  - ")
		builder.WriteString(err.Error())
		builder.WriteString("\n")
	}
	return builder.String()
}

func (v *ValidationErrors) Add(field, value, message string) {
	v.Errors = append(v.Errors, &ConfigError{
		Field:   field,
		Value:   value,
		Message: message,
	})
}

func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// Loader は設定のローダーインターフェース
type Loader interface {
	Load() (Config, error)
}

// EnvLoader は環境変数から設定を読み込むローダー
type EnvLoader struct {
	envFile string
}

// NewEnvLoader は新しいEnvLoaderを作成する
func NewEnvLoader(envFile ...string) *EnvLoader {
	file := ".env"
	if len(envFile) > 0 && envFile[0] != "" {
		file = envFile[0]
	}
	return &EnvLoader{envFile: file}
}

// Load は環境変数から設定を読み込む
func (l *EnvLoader) Load() (Config, error) {
	// .envファイルの読み込み（エラーは無視）
	_ = godotenv.Load(l.envFile)

	config := &appConfig{
		environment: getEnvOrDefault("APP_ENV", "development"),
		version:     getEnvOrDefault("APP_VERSION", "dev"),
	}

	var validationErrors ValidationErrors

	// Database設定の構築
	dbConfig, err := l.loadDatabaseConfig()
	if err != nil {
		if ve, ok := err.(*ValidationErrors); ok {
			validationErrors.Errors = append(validationErrors.Errors, ve.Errors...)
		} else {
			return nil, err
		}
	}
	config.database = dbConfig

	// JWT設定の構築
	jwtConfig, err := l.loadJWTConfig()
	if err != nil {
		if ve, ok := err.(*ValidationErrors); ok {
			validationErrors.Errors = append(validationErrors.Errors, ve.Errors...)
		} else {
			return nil, err
		}
	}
	config.jwt = jwtConfig

	// Server設定の構築
	serverConfig, err := l.loadServerConfig()
	if err != nil {
		if ve, ok := err.(*ValidationErrors); ok {
			validationErrors.Errors = append(validationErrors.Errors, ve.Errors...)
		} else {
			return nil, err
		}
	}
	config.server = serverConfig

	// Log設定の構築
	logConfig, err := l.loadLogConfig(config.environment)
	if err != nil {
		if ve, ok := err.(*ValidationErrors); ok {
			validationErrors.Errors = append(validationErrors.Errors, ve.Errors...)
		} else {
			return nil, err
		}
	}
	config.log = logConfig

	if validationErrors.HasErrors() {
		return nil, &validationErrors
	}

	// 最終バリデーション
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (l *EnvLoader) loadDatabaseConfig() (databaseConfig, error) {
	var errors ValidationErrors

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		errors.Add("DATABASE_URL", "", "required")
	}

	maxConnections := getEnvAsIntOrDefault("DB_MAX_CONNECTIONS", 25)
	if maxConnections <= 0 {
		errors.Add("DB_MAX_CONNECTIONS", strconv.Itoa(maxConnections), "must be positive")
	}

	maxIdleTime := getEnvAsDurationOrDefault("DB_MAX_IDLE_TIME", 15*time.Minute)
	if maxIdleTime < 0 {
		errors.Add("DB_MAX_IDLE_TIME", maxIdleTime.String(), "must be non-negative")
	}

	connMaxLifetime := getEnvAsDurationOrDefault("DB_CONN_MAX_LIFETIME", time.Hour)
	if connMaxLifetime < 0 {
		errors.Add("DB_CONN_MAX_LIFETIME", connMaxLifetime.String(), "must be non-negative")
	}

	if errors.HasErrors() {
		return databaseConfig{}, &errors
	}

	return databaseConfig{
		url:             url,
		maxConnections:  maxConnections,
		maxIdleTime:     maxIdleTime,
		connMaxLifetime: connMaxLifetime,
	}, nil
}

func (l *EnvLoader) loadJWTConfig() (jwtConfig, error) {
	var errors ValidationErrors

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		errors.Add("JWT_SECRET", "", "required")
	} else if len(secret) < 32 {
		errors.Add("JWT_SECRET", "***", "must be at least 32 characters")
	}

	accessExp := getEnvAsDurationOrDefault("JWT_ACCESS_TOKEN_EXPIRATION", 15*time.Minute)
	if accessExp <= 0 {
		errors.Add("JWT_ACCESS_TOKEN_EXPIRATION", accessExp.String(), "must be positive")
	}

	refreshExp := getEnvAsDurationOrDefault("JWT_REFRESH_TOKEN_EXPIRATION", 7*24*time.Hour)
	if refreshExp <= 0 {
		errors.Add("JWT_REFRESH_TOKEN_EXPIRATION", refreshExp.String(), "must be positive")
	}

	issuer := getEnvOrDefault("JWT_ISSUER", "travel-api")

	if errors.HasErrors() {
		return jwtConfig{}, &errors
	}

	return jwtConfig{
		secret:                 secret,
		accessTokenExpiration:  accessExp,
		refreshTokenExpiration: refreshExp,
		issuer:                 issuer,
	}, nil
}

func (l *EnvLoader) loadServerConfig() (serverConfig, error) {
	var errors ValidationErrors

	port := getEnvOrDefault("PORT", "8080")
	host := getEnvOrDefault("HOST", "0.0.0.0")

	readTimeout := getEnvAsDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second)
	if readTimeout < 0 {
		errors.Add("SERVER_READ_TIMEOUT", readTimeout.String(), "must be non-negative")
	}

	writeTimeout := getEnvAsDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second)
	if writeTimeout < 0 {
		errors.Add("SERVER_WRITE_TIMEOUT", writeTimeout.String(), "must be non-negative")
	}

	idleTimeout := getEnvAsDurationOrDefault("SERVER_IDLE_TIMEOUT", 60*time.Second)
	if idleTimeout < 0 {
		errors.Add("SERVER_IDLE_TIMEOUT", idleTimeout.String(), "must be non-negative")
	}

	shutdownTimeout := getEnvAsDurationOrDefault("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second)
	if shutdownTimeout < 0 {
		errors.Add("SERVER_SHUTDOWN_TIMEOUT", shutdownTimeout.String(), "must be non-negative")
	}

	if errors.HasErrors() {
		return serverConfig{}, &errors
	}

	return serverConfig{
		port:            port,
		host:            host,
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		idleTimeout:     idleTimeout,
		shutdownTimeout: shutdownTimeout,
	}, nil
}

func (l *EnvLoader) loadLogConfig(environment string) (logConfig, error) {
	var errors ValidationErrors

	levelStr := getEnvOrDefault("LOG_LEVEL", "info")
	level, err := parseLogLevel(levelStr)
	if err != nil {
		errors.Add("LOG_LEVEL", levelStr, err.Error())
	}

	format := getEnvOrDefault("LOG_FORMAT", "json")
	validFormats := []string{"json", "text"}
	if !contains(validFormats, format) {
		errors.Add("LOG_FORMAT", format, fmt.Sprintf("must be one of: %s", strings.Join(validFormats, ", ")))
	}

	// 開発環境では詳細ログを出力
	addSource := environment == "development" || level == slog.LevelDebug

	if errors.HasErrors() {
		return logConfig{}, &errors
	}

	return logConfig{
		level:     level,
		format:    format,
		addSource: addSource,
	}, nil
}

// appConfig のバリデーションメソッド
func (c appConfig) Validate() error {
	var errors ValidationErrors

	// 追加のクロスフィールドバリデーション
	if c.jwt.AccessTokenExpiration() > c.jwt.RefreshTokenExpiration() {
		errors.Add("JWT_ACCESS_TOKEN_EXPIRATION", "", "must not be greater than refresh token expiration")
	}

	if errors.HasErrors() {
		return &errors
	}

	return nil
}

// ヘルパー関数
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		slog.Warn("Invalid integer value for environment variable, using default",
			"key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		slog.Warn("Invalid duration value for environment variable, using default",
			"key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}

func parseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", level)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// パッケージレベルの便利関数
var defaultLoader = NewEnvLoader()

// Load はデフォルトローダーを使用して設定を読み込む
func Load() (Config, error) {
	return defaultLoader.Load()
}

// LoadWithFile は指定された.envファイルを使用して設定を読み込む
func LoadWithFile(envFile string) (Config, error) {
	loader := NewEnvLoader(envFile)
	return loader.Load()
}

// AccessTokenExpiration は、アクセストークンの有効期限を環境変数から取得します。
// 環境変数が設定されていない場合は、デフォルトの24時間を返します。
func AccessTokenExpiration() time.Duration {
	_ = godotenv.Load()
	expStr := os.Getenv("ACCESS_TOKEN_EXPIRATION_HOURS")
	if expStr == "" {
		return 24 * time.Hour // デフォルト値
	}
	hours, err := strconv.Atoi(expStr)
	if err != nil {
		slog.Warn("Invalid ACCESS_TOKEN_EXPIRATION_HOURS, using default", "value", expStr)
		return 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

// RefreshTokenExpiration は、リフレッシュトークンの有効期限を環境変数から取得します。
// 環境変数が設定されていない場合は、デフォルトの7日間を返します。
func RefreshTokenExpiration() time.Duration {
	_ = godotenv.Load()
	expStr := os.Getenv("REFRESH_TOKEN_EXPIRATION_DAYS")
	if expStr == "" {
		return 7 * 24 * time.Hour // デフォルト値
	}
	days, err := strconv.Atoi(expStr)
	if err != nil {
		slog.Warn("Invalid REFRESH_TOKEN_EXPIRATION_DAYS, using default", "value", expStr)
		return 7 * 24 * time.Hour
	}
	return time.Duration(days) * 24 * time.Hour
}
