package config

import (
	"log"
	"path/filepath"
	"sync"
	"sync/atomic"
	"yunyez/internal/common/tools"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// 热加载配置
// 包含特定环境配置和公共配置
// 特定环境配置文件路径：/configs/<environment>/config-name.yaml
// 公共配置文件路径：/configs/config-name.yaml

// 基本配置文件
var (
	BASE_CONFIG_FILE   = "configs/config.yaml" // 基本配置文件
	BASE_CONFIG_FOLDER = "configs/"            // 基本配置文件目录
)

// configHolder 用于原子地保存配置对象
type configHolder struct {
	data atomic.Value // 保存 *viper.Viper
}

var (
	cfgHolder = &configHolder{}       // 全局配置持有者
	watch     = make(map[string]bool) // 配置文件监控列表
	once      sync.Once

	commonConfigFiles = []string{
		"device.yaml",
	}
	envConfigFiles = []string{
		"database.yaml",
	}
)

func init() {
	_ = initConfig()
}

// Init 初始化配置文件 供外部显示调用
func Init() error {
	var err error
	once.Do(func() {
		err = initConfig()
	})
	return err
}

// initConfig 初始化配置文件
func initConfig() error {
	// 创建新的 viper 实例
	newViper := viper.New()

	// // 当前工作目录
	// wd, err := os.Getwd()
	// if err != nil {
	// 	return err
	// }

	// 获取项目根目录
	wd := tools.GetRootDir()
	defaultConfig := filepath.Join(wd, BASE_CONFIG_FILE)
	newViper.SetConfigFile(defaultConfig)
	watchConfig(newViper, defaultConfig)
	if err := newViper.MergeInConfig(); err != nil {
		return err
	}

	// 获取环境变量
	environment := newViper.GetString("app.env")
	if environment == "" {
		log.Fatalf("env is not set in config file: %s", filepath.Join(wd, BASE_CONFIG_FILE))
	}

	for _, file := range commonConfigFiles {
		path := filepath.Join(wd, BASE_CONFIG_FOLDER, file)
		log.Printf("common config file path: %s", path)
		newViper.SetConfigFile(path)
		watchConfig(newViper, path)
		if err := newViper.MergeInConfig(); err != nil {
			return err
		}
	}

	for _, file := range envConfigFiles {
		path := filepath.Join(wd, BASE_CONFIG_FOLDER, environment, file)
		log.Printf("special config file path: %s", path)
		newViper.SetConfigFile(path)
		watchConfig(newViper, path)
		if err := newViper.MergeInConfig(); err != nil {
			return err
		}
	}

	// 原子替换配置对象
	cfgHolder.data.Store(newViper)
	log.Printf("init config success...")

	return nil
}

// watchConfig 监控配置文件变化
// 使用 viper.WatchConfig 监控配置文件变化
// 当配置文件变化时，会调用 viper.OnConfigChange 注册的回调函数处理变更
func watchConfig(v *viper.Viper, path string) {
	// 检查配置路径是否已在监控列表中
	if _, ok := watch[path]; ok {
		return
	}
	watch[path] = true

	// 注册配置变更回调函数
	v.OnConfigChange(func(e fsnotify.Event) {
		// 打印配置变更事件信息
		log.Printf("config file %s changed", e.Name)
		// 重新加载配置
		if err := initConfig(); err != nil {
			log.Printf("Error reading config file: %s", err)
		} else {
			log.Printf("config file %s reloaded...", e.Name)
		}
	})
	v.WatchConfig()
}

// getViper 获取当前的 viper 实例
func getViper() *viper.Viper {
	v := cfgHolder.data.Load()
	if v != nil && v.(*viper.Viper) != nil {
		return v.(*viper.Viper)
	}

	once.Do(func() {
		_ = initConfig()
	})
	v = cfgHolder.data.Load()
	if v == nil || v.(*viper.Viper) == nil {
		log.Fatalf("viper instance is nil")
	}
	
	return v.(*viper.Viper)
}

// ======= 配置文件相关函数 =======

// GetString 获取字符串配置值
func GetString(key string) string {
	return getViper().GetString(key)
}

// GetStringWithDefault 获取字符串配置值，若为空则返回默认值
func GetStringWithDefault(key string, defaultValue string) string {
	val := getViper().GetString(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// GetInt 获取整数配置值
func GetInt(key string) int {
	return getViper().GetInt(key)
} 

// GetIntWithDefault 获取整数配置值，若为空则返回默认值
func GetIntWithDefault(key string, defaultValue int) int {
	val := getViper().GetInt(key)
	if val == 0 {
		return defaultValue
	}
	return val
}

// GetBool 获取布尔配置值
func GetBool(key string) bool {
	val := getViper().GetBool(key)
	return val
}

// GetList 获取字符串列表配置值
func GetList(key string) []string {
	return getViper().GetStringSlice(key)
}
