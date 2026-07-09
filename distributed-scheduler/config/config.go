package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	Etcd   EtcdConfig   `yaml:"etcd"`
	Admin  AdminConfig  `yaml:"admin"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type EtcdConfig struct {
	Endpoints []string `yaml:"endpoints"`
}

type AdminConfig struct {
	AccessToken            string `yaml:"access_token"`
	ExecutorHeartbeatTTL   int    `yaml:"executor_heartbeat_ttl"`
	ExecutorDiscoverPrefix string `yaml:"executor_discover_prefix"`
}

// Global 全局配置实例
var Global Config

// Load 从 YAML 文件加载配置，并允许环境变量覆盖关键字段
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	if err := yaml.Unmarshal(data, &Global); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}
	applyEnv()
	// 配置兜底
	if Global.Admin.ExecutorDiscoverPrefix == "" {
		Global.Admin.ExecutorDiscoverPrefix = "dscheduler/executors/"
	}
	if Global.Admin.ExecutorHeartbeatTTL == 0 {
		Global.Admin.ExecutorHeartbeatTTL = 10
	}
	return nil
}

func applyEnv() {
	if v := os.Getenv("MYSQL_PASSWORD"); v != "" {
		Global.MySQL.Password = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			Global.Server.Port = p
		}
	}
}
