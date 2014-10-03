package config

import (
	logging "github.com/op/go-logging"
	"github.com/viert/properties"
)

type Config struct {
	Host           string
	Port           int
	IndexDirectory string
	SyncBufferSize int
	GCPercent      int
	MaxCores       int
	MaxThreads     int
}

var (
	log           *logging.Logger = logging.MustGetLogger("metricsearch")
	defaultConfig *Config         = &Config{
		Host:           "",
		Port:           7000,
		IndexDirectory: "/var/lib/metricsearch/index",
		SyncBufferSize: 1000,
		GCPercent:      100,
		MaxCores:       8,
		MaxThreads:     10000,
	}
)

func Load(filename string) *Config {
	props, err := properties.Load(filename)
	if err != nil {
		log.Error(err.Error())
		log.Notice("Using configuration defaults")
		return defaultConfig
	}
	config := new(Config)
	config.Host, err = props.GetString("main.host")
	config.Port, err = props.GetInt("main.port")
	if err != nil {
		config.Port = defaultConfig.Port
	}
	config.IndexDirectory, err = props.GetString("main.index_directory")
	if err != nil {
		config.IndexDirectory = defaultConfig.IndexDirectory
	}
	config.SyncBufferSize, err = props.GetInt("main.sync_buffer_size")
	if err != nil {
		config.SyncBufferSize = defaultConfig.SyncBufferSize
	}
	config.GCPercent, err = props.GetInt("runtime.gc_percent")
	if err != nil {
		config.GCPercent = defaultConfig.GCPercent
	}
	config.MaxCores, err = props.GetInt("runtime.max_cores")
	if err != nil {
		config.MaxCores = defaultConfig.MaxCores
	}
	config.MaxThreads, err = props.GetInt("runtime.max_threads")
	if err != nil {
		config.MaxThreads = defaultConfig.MaxThreads
	}
	return config
}
