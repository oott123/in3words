// Package config 定义服务配置，并通过 viper 从配置文件与环境变量加载。
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const defaultCacheSize = 1024

// EnvPrefix 是环境变量前缀，例如 db.password 对应 I3W_DB_PASSWORD。
const EnvPrefix = "I3W"

// Config 是服务的全部配置。
type Config struct {
	Listen string `mapstructure:"listen"`
	Log    Log    `mapstructure:"log"`
	DB     DB     `mapstructure:"db"`
	Render Render `mapstructure:"render"`
}

// Log 是日志配置。
type Log struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// DB 对应 wp-config.php 中的数据库设置。
type DB struct {
	// Host 与 DB_HOST 相同，可以是 "host"、"host:port"，或 "localhost:/path/to/mysqld.sock"。
	Host        string `mapstructure:"host"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Name        string `mapstructure:"name"`
	Charset     string `mapstructure:"charset"`
	TablePrefix string `mapstructure:"table_prefix"`
}

// Render 是内容渲染相关的配置，对应 WP 中无法从数据库读到的环境信息。
type Render struct {
	// HTML5Caption 对应主题是否声明了 add_theme_support('html5', ['caption'])。
	HTML5Caption bool `mapstructure:"html5_caption"`
	// UploadsBaseURL 对应 wp_get_upload_dir()['baseurl']，为空时按 WP 默认规则从站点选项推算；
	// wp-config.php 中定义了 WP_CONTENT_URL 或 UPLOADS 时需要手动设置。
	UploadsBaseURL string `mapstructure:"uploads_base_url"`
	// CacheSize 是渲染结果缓存的条目数，0 表示不缓存。
	CacheSize int `mapstructure:"cache_size"`
}

// SetDefaults 写入各配置项的默认值。
func SetDefaults(v *viper.Viper) {
	v.SetDefault("listen", ":8080")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("db.host", "localhost:3306")
	v.SetDefault("db.user", "")
	v.SetDefault("db.password", "")
	v.SetDefault("db.name", "wordpress")
	v.SetDefault("db.charset", "utf8mb4")
	v.SetDefault("db.table_prefix", "wp_")
	v.SetDefault("render.html5_caption", true)
	v.SetDefault("render.uploads_base_url", "")
	v.SetDefault("render.cache_size", defaultCacheSize)
}

// Load 读取配置文件（file 为空时只用默认值和环境变量）并解析为 Config。
func Load(v *viper.Viper, file string) (*Config, error) {
	SetDefaults(v)
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if file != "" {
		v.SetConfigFile(file)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	var errs []error
	if c.DB.Name == "" {
		errs = append(errs, errors.New("db.name is required"))
	}
	if c.DB.TablePrefix == "" {
		errs = append(errs, errors.New("db.table_prefix is required"))
	}
	if c.Render.CacheSize < 0 {
		errs = append(errs, fmt.Errorf("render.cache_size %d is negative", c.Render.CacheSize))
	}
	return errors.Join(errs...)
}
