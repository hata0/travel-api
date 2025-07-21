package injector

// Config は設定値を保持する
type Config struct {
	JWTSecret string
}

// NewConfig は設定を初期化する
func NewConfig(jwtSecret string) *Config {
	return &Config{
		JWTSecret: jwtSecret,
	}
}
