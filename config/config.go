package config

type Config struct {
	App   AppConfig   `yaml:"app"`
	MySQL MySQLConfig `yaml:"mysql"`
	Redis RedisConfig `yaml:"redis"`
	JWT   JWTConfig   `yaml:"jwt"`
	AI    AIConfig    `yaml:"ai"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type AIConfig struct {
	BaseURL        string `yaml:"base_url"`
	APIKey         string `yaml:"api_key"`
	EmbeddingModel string `yaml:"embedding_model"`
	ChatModel      string `yaml:"chat_model"`
}