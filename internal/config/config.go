package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	ZLMediaKit ZLMediaKitConfig `mapstructure:"zlmediakit"`
	Device     DeviceConfig     `mapstructure:"device"`
	Log        LogConfig        `mapstructure:"log"`
}

type ServerConfig struct {
	ID       string `mapstructure:"id"`
	Domain   string `mapstructure:"domain"`
	Host     string `mapstructure:"host"`
	SIPPort  int    `mapstructure:"sip_port"`
	HTTPPort int    `mapstructure:"http_port"`
	Password string `mapstructure:"password"`
}

type ZLMediaKitConfig struct {
	Host     string `mapstructure:"host"`
	HTTPPort int    `mapstructure:"http_port"`
	Secret   string `mapstructure:"secret"`
}

type DeviceConfig struct {
	HeartbeatTimeout int `mapstructure:"heartbeat_timeout"`
	RegisterExpire   int `mapstructure:"register_expire"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

var GlobalConfig *Config

func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

func (c *Config) SIPAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.SIPPort)
}

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.HTTPPort)
}

func (c *Config) ZLMediaKitAPIBase() string {
	return fmt.Sprintf("http://%s:%d", c.ZLMediaKit.Host, c.ZLMediaKit.HTTPPort)
}
