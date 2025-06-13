package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	Cfg         *Config
	configMutex = &sync.RWMutex{}
)

type ServerInfo struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Msg struct {
	MsgExpireTime      int64         `mapstructure:"msg_expire_time"`
	HeartbeatInterval  time.Duration `mapstructure:"heartbeat_interval"`
	HeartbeatTimeout   int64         `mapstructure:"heartbeat_timeout"`
	HeartbeatCheckTime time.Duration `mapstructure:"heartbeat_check_time"`
}

type Config struct {
	SrvInfo ServerInfo `mapstructure:"srvInfo"`
	Msg     Msg        `mapstructure:"msg"`
}

func initBase(configPath string, setDefaultFunc func(v *viper.Viper)) *viper.Viper {
	v := viper.New()

	// 基础配置
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 环境变量支持
	v.AutomaticEnv()

	// 默认值设置
	setDefaultFunc(v)

	return v
}

func getCfgPath() string {
	configFile := "config/dev.config.yaml"
	env := os.Getenv("TCPSOCKET_ENV")
	if strings.ToLower(env) == "prod" {
		configFile = "config/config.yaml"
	} else if strings.ToLower(env) == "test" {
		configFile = "config/test.config.yaml"
	}
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, configFile)
}
func Init() {
	v := viper.New()

	configPath := getCfgPath()
	// 基础配置
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 环境变量支持
	v.AutomaticEnv()
	v.SetEnvPrefix("KIN")

	// 设置默认值
	setDefault(v)

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("加载配置文件异常: %w", err))
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("解析配置文件异常: %w", err))
	}

	if err := validateCfg(&cfg); err != nil {
		panic(fmt.Errorf("配置文件校验不通过: %w", err))
	}

	// 设置全局配置
	configMutex.Lock()
	Cfg = &cfg
	configMutex.Unlock()

	// 开启热重载
	enableReload(v)

}

func setDefault(v *viper.Viper) {
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", 8000)
	v.SetDefault("msg.msg_expire_time", 60)
	v.SetDefault("msg.heartbeat_interval", 5)
	v.SetDefault("msg.heartbeat_timeout", 60)
	v.SetDefault("msg.heartbeat_check_time", 15)
}

// ValidateCfg 配置校验
func validateCfg(cfg *Config) error {
	return nil
}

func enableReload(v *viper.Viper) {
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)

		var newConfig Config
		if err := v.Unmarshal(&newConfig); err != nil {
			fmt.Printf("failed to reload config: %v\n", err)
			return
		}

		if err := validateCfg(&newConfig); err != nil {
			fmt.Printf("invalid config: %v\n", err)
			return
		}

		configMutex.Lock()
		Cfg = &newConfig
		configMutex.Unlock()
		fmt.Println("Configuration reloaded successfully")
	})
}

// Get 获取全局配置
func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return Cfg
}
