package config

import (
	logging "github.com/op/go-logging"
	"github.com/viert/properties"
	"strings"
)

type Config struct {
	Host              string
	Port              int
	IndexDirectory    string
	SyncBufferSize    int
	GCPercent         int
	MaxCores          int
	MaxThreads        int
	LogLevel          logging.Level
	Log               string
	SelfMonitor       bool
	SelfMonitorPrefix string
	ValidateTokens    bool
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
		LogLevel:       logging.DEBUG,
		Log:            "",
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
	if err != nil {
		config.Host = defaultConfig.Host
	}
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
	config.Log, err = props.GetString("main.log")
	if err != nil {
		config.Log = defaultConfig.Log
	}
	validateTokens, err := props.GetString("main.validate_tokens")
	if err == nil {
		switch strings.ToLower(validateTokens) {
		case "on":
			fallthrough
		case "1":
			fallthrough
		case "yes":
			fallthrough
		case "true":
			config.ValidateTokens = true
		default:
			config.ValidateTokens = false
		}
	}

	logLevel, err := props.GetString("main.log_level")
	if err != nil {
		config.LogLevel = defaultConfig.LogLevel
	} else {
		switch strings.ToLower(logLevel) {
		case "debug":
			config.LogLevel = logging.DEBUG
		case "error":
			config.LogLevel = logging.ERROR
		case "info":
			config.LogLevel = logging.INFO
		case "critical":
			config.LogLevel = logging.CRITICAL
		case "notice":
			config.LogLevel = logging.NOTICE
		case "warning":
			config.LogLevel = logging.WARNING
		default:
			config.LogLevel = defaultConfig.LogLevel
		}
	}
	dsSync, err := props.GetString("main.no_sync")
	if err == nil {
		dsSync = strings.ToLower(dsSync)
		if dsSync == "true" {
			config.SyncBufferSize = -1
		}
	}
	selfMonitor, err := props.GetString("main.self_monitor")
	if err == nil {
		switch strings.ToLower(selfMonitor) {
		case "on":
			fallthrough
		case "1":
			fallthrough
		case "yes":
			fallthrough
		case "true":
			config.SelfMonitor = true
		default:
			config.SelfMonitor = false
		}
	}
	selfMonitorPrefix, err := props.GetString("main.self_monitor_prefix")
	if err != nil {
		config.SelfMonitorPrefix = ""
	} else {
		if strings.HasSuffix(selfMonitorPrefix, ".") {
			selfMonitorPrefix = selfMonitorPrefix[0:len(selfMonitorPrefix)]
		}
		config.SelfMonitorPrefix = selfMonitorPrefix
	}
	return config
}
