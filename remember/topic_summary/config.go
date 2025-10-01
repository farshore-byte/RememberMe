package topic_summary

import (
	"log"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host     string
	Port     int
	DB       int
	Password string
	SSL      bool
}

type MongoConfig struct {
	URI string
	TLS bool
	DB  string
}

type LLMConfig struct {
	ServiceProvider string
	APIKey          string `mapstructure:"api_key"`
	BaseURL         string `mapstructure:"base_url"`
	ModelID         string `mapstructure:"model_id"`
	Temperature     float64
	TopP            float64 `mapstructure:"top_p"`
	MaxNewTokens    int     `mapstructure:"max_new_tokens"`
}

type AuthConfig struct {
	Token string
}

type FeishuConfig struct {
	Webhook string
}

type ServerConfig struct {
	SessionMessages int `mapstructure:"session_messages"`
	UserPortrait    int `mapstructure:"user_poritrait"`
	TopicSummary    int `mapstructure:"topic_summary"`
	ChatEvent       int `mapstructure:"chat_event"`
	Main            int `mapstructure:"main"`
}
type AppConfig struct {
	Redis   RedisConfig
	MongoDB MongoConfig `mapstructure:"mongodb"`
	LLM     LLMConfig
	Feishu  FeishuConfig
	Auth    AuthConfig
	Server  ServerConfig
}

var Config AppConfig

func init() {
	viper.SetConfigName("config") // 不要带 .yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // 根目录
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}
	log.Printf("init config success")

}
