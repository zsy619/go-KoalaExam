package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 整体配置
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Log       LogConfig       `mapstructure:"log"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	Upload    UploadConfig    `mapstructure:"upload"`
	AntiCheat AntiCheatConfig `mapstructure:"anti_cheat"`
}

type AppConfig struct {
	Name          string `mapstructure:"name"`
	Mode          string `mapstructure:"mode"`
	Version       string `mapstructure:"version"`
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	ReadTimeout   int    `mapstructure:"read_timeout"`
	WriteTimeout  int    `mapstructure:"write_timeout"`
	EnableNetpoll bool   `mapstructure:"enable_netpoll"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

type MySQLConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`
	LogLevel        string `mapstructure:"log_level"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	Issuer        string `mapstructure:"issuer"`
	AccessExpire  int    `mapstructure:"access_expire"`
	RefreshExpire int    `mapstructure:"refresh_expire"`
	TokenHeader   string `mapstructure:"token_header"`
	TokenPrefix   string `mapstructure:"token_prefix"`
}

type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	QPS     int  `mapstructure:"qps"`
	Burst   int  `mapstructure:"burst"`
	Expire  int  `mapstructure:"expire"`
}

type UploadConfig struct {
	Driver   string   `mapstructure:"driver"`
	Path     string   `mapstructure:"path"`
	BaseURL  string   `mapstructure:"base_url"`
	MaxSize  int      `mapstructure:"max_size"`
	AllowExt []string `mapstructure:"allow_ext"`
}

type AntiCheatConfig struct {
	EnableFullscreen      bool `mapstructure:"enable_fullscreen"`
	EnableTabSwitchDetect bool `mapstructure:"enable_tab_switch_detect"`
	MaxTabSwitchCount     int  `mapstructure:"max_tab_switch_count"`
	EnableCopyPasteBlock  bool `mapstructure:"enable_copy_paste_block"`
	ShuffleQuestions      bool `mapstructure:"shuffle_questions"`
	ShuffleOptions        bool `mapstructure:"shuffle_options"`
	SaveProgressInterval  int  `mapstructure:"save_progress_interval"`
}

// LoadConfig 加载配置
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvPrefix("KOALA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}
	return &c, nil
}

// LoadConfigByEnv 根据环境加载配置（开发/生产）
func LoadConfigByEnv(env string) (*Config, error) {
	if env == "" {
		env = "dev"
	}
	file := fmt.Sprintf("configs/config.%s.yaml", env)
	if env == "dev" {
		file = "configs/config.yaml"
	}
	return LoadConfig(file)
}
