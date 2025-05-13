package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 结构体定义了所有配置项
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	DeepSeek DeepSeekConfig `mapstructure:"deepseek"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port            int    `mapstructure:"port"`
	Host            string `mapstructure:"host"`
	RateLimit       int    `mapstructure:"rate_limit"`       // 每分钟请求限制
	RateLimitBurst  int    `mapstructure:"rate_limit_burst"` // 突发请求限制
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"` // 优雅关闭超时时间（秒）
}

// DeepSeekConfig DeepSeek API配置
type DeepSeekConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	BaseURL     string  `mapstructure:"base_url"`
	Temperature float64 `mapstructure:"temperature"`
	Timeout     int     `mapstructure:"timeout"` // API调用超时时间（秒）
}

// LogConfig 日志配置
type LogConfig struct {
	Enabled    bool   `mapstructure:"enabled"`     // 是否启用文件日志
	Level      string `mapstructure:"level"`       // 日志级别
	File       string `mapstructure:"file"`        // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 单个日志文件最大尺寸（MB）
	MaxBackups int    `mapstructure:"max_backups"` // 最大保留的旧日志文件数
	MaxAge     int    `mapstructure:"max_age"`     // 旧日志文件保留的最大天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩旧日志文件
}

var cfg *Config

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("获取可执行文件路径失败: %v", err)
	}
	execDir := filepath.Dir(execPath)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(execDir)
	viper.AddConfigPath(filepath.Join(execDir, "config"))
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	setDefaults(execDir)

	// 读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AGENT_FORGE")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在时使用默认配置
			cfg = &Config{}
			if err := viper.Unmarshal(cfg); err != nil {
				return nil, fmt.Errorf("unmarshal default config failed: %v", err)
			}
			fmt.Fprintf(os.Stderr, "\n警告: 配置文件不存在，使用默认配置\n")
			fmt.Fprintf(os.Stderr, "请在以下位置创建配置文件：\n")
			fmt.Fprintf(os.Stderr, "- %s/config.yaml\n", execDir)
			fmt.Fprintf(os.Stderr, "- %s/config/config.yaml\n\n", execDir)
			fmt.Fprintf(os.Stderr, "默认配置如下：\n%v\n", viper.AllSettings())
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	} else {
		cfg = &Config{}
		if err := viper.Unmarshal(cfg); err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %v", err)
		}
	}

	// 环境变量覆盖
	if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
		cfg.DeepSeek.APIKey = apiKey
	}

	return cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults(execDir string) {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.rate_limit", 60)
	viper.SetDefault("server.rate_limit_burst", 10)
	viper.SetDefault("server.shutdown_timeout", 30)

	viper.SetDefault("deepseek.base_url", "https://api.deepseek.com")
	viper.SetDefault("deepseek.temperature", 0.7)
	viper.SetDefault("deepseek.timeout", 30)

	// 使用绝对路径设置日志文件路径
	defaultLogPath := filepath.Join(execDir, "logs", "agent-forge.log")
	viper.SetDefault("log.enabled", false) // 默认关闭文件日志
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", defaultLogPath)
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 28)
	viper.SetDefault("log.compress", true)
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	if cfg == nil {
		cfg, _ = LoadConfig()
	}
	return cfg
}
