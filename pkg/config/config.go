// Package config has specific primitives for loading application configurations.
//
// Primitives:
//   - Application should have struct for containing configuration. E.g. refer
//     propeller/internal/config/config.go file.
//   - Application should have a directory holding default file and environment
//     specific file. E.g. refer propeller/configs/* directory.
//
// Usage:
//   - E.g. NewDefaultConfig().Load("dev", &config), where config is a struct
//     where configuration gets unmarshalled into.
package config

import (
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Default options for configuration loading.
const (
	DefaultConfigType     = "toml"
	DefaultConfigDir      = "./config"
	DefaultConfigFileName = "propeller"
	FilePath              = "PROPELLER_CONFIG_FILE_PATH"
)

// Options is config options.
type Options struct {
	configType            string
	configPath            string
	defaultConfigFileName string
}

// Config is a wrapper over a underlying config loader implementation.
type Config struct {
	opts  Options
	viper *viper.Viper
}

// NewDefaultOptions returns default options.
// DISCLAIMER: This function is a bit hacky
// This function expects an env $WORKDIR to
// be set and reads configs from $WORKDIR/configs.
// If $WORKDIR is not set. It uses the absolute path wrt
// the location of this file (config.go) to set configPath
// to 2 levels up in viper (../../configs).
// This function breaks if :
// 1. $WORKDIR is set and configs dir not present in $WORKDIR
// 2. $WORKDIR is not set and ../../configs is not present
// 3. $WORKDIR is not set and runtime absolute path of configs
// is different than build time path as runtime.Caller() evaluates
// only at build time
func NewDefaultOptions() Options {
	var configPath string
	filePath := os.Getenv(FilePath)
	if filePath != "" {
		configPath = path.Join(filePath)
	} else {
		_, thisFile, _, _ := runtime.Caller(1)
		configPath = path.Join(path.Dir(thisFile), "../../"+DefaultConfigDir)
	}
	return NewOptions(DefaultConfigType, configPath, DefaultConfigFileName)
}

// NewOptions returns new Options struct.
func NewOptions(configType string, configPath string, defaultConfigFileName string) Options {
	return Options{configType, configPath, defaultConfigFileName}
}

// NewDefaultConfig returns new config struct with default options.
func NewDefaultConfig() *Config {
	return NewConfig(NewDefaultOptions())
}

// NewConfig returns new config struct.
func NewConfig(opts Options) *Config {
	return &Config{opts, viper.New()}
}

// Load reads environment specific configurations and along with the defaults
// unmarshalls into config.
func (c *Config) Load(fileName string, config interface{}) error {
	return c.loadByConfigName(fileName, config)
}

// loadByConfigName reads configuration from file and unmarshalls into config.
func (c *Config) loadByConfigName(configName string, config interface{}) error {
	c.viper.SetEnvPrefix(strings.ToUpper("PROPELLER"))
	c.viper.SetConfigName(configName)
	c.viper.SetConfigType(c.opts.configType)
	c.viper.AddConfigPath(c.opts.configPath)
	c.viper.AutomaticEnv()
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := c.viper.ReadInConfig(); err != nil {
		return err
	}
	return c.viper.Unmarshal(config)
}
