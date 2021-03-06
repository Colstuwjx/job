// Copyright Project Harbor Authors. All rights reserved.

// Package config provides functions to handle the configurations of job service.
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/Colstuwjx/job/utils"
)

const (
	jobServiceProtocol            = "JOB_SERVICE_PROTOCOL"
	jobServicePort                = "JOB_SERVICE_PORT"
	jobServiceHTTPCert            = "JOB_SERVICE_HTTPS_CERT"
	jobServiceHTTPKey             = "JOB_SERVICE_HTTPS_KEY"
	jobServiceWorkerPoolBackend   = "JOB_SERVICE_POOL_BACKEND"
	jobServiceWorkers             = "JOB_SERVICE_POOL_WORKERS"
	jobServiceRedisURL            = "JOB_SERVICE_POOL_REDIS_URL"
	jobServiceRedisNamespace      = "JOB_SERVICE_POOL_REDIS_NAMESPACE"
	jobServiceLoggerBasePath      = "JOB_SERVICE_LOGGER_BASE_PATH"
	jobServiceLoggerLevel         = "JOB_SERVICE_LOGGER_LEVEL"
	jobServiceLoggerArchivePeriod = "JOB_SERVICE_LOGGER_ARCHIVE_PERIOD"
	jobServiceAuthSecret          = "JOBSERVICE_SECRET"

	// JobServiceProtocolHTTPS points to the 'https' protocol
	JobServiceProtocolHTTPS = "https"

	// JobServiceProtocolHTTP points to the 'http' protocol
	JobServiceProtocolHTTP = "http"

	// JobServicePoolBackendRedis represents redis backend
	JobServicePoolBackendRedis = "redis"

	// secret of UI
	uiAuthSecret = "CORE_SECRET"

	// redis protocol schema
	redisSchema = "redis://"
)

// DefaultConfig is the default configuration reference
var DefaultConfig = &Configuration{}

// Configuration loads and keeps the related configuration items of job service.
type Configuration struct {
	// Protocol server listening on: https/http
	Protocol string `yaml:"protocol"`

	// Server listening port
	Port uint `yaml:"port"`

	// Additional config when using https
	HTTPSConfig *HTTPSConfig `yaml:"https_config,omitempty"`

	// Configurations of worker pool
	PoolConfig *PoolConfig `yaml:"worker_pool,omitempty"`

	// Logger configurations
	LoggerConfig *LoggerConfig `yaml:"logger,omitempty"`
}

// HTTPSConfig keeps additional configurations when using https protocol
type HTTPSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

// RedisPoolConfig keeps redis pool info.
type RedisPoolConfig struct {
	RedisURL  string `yaml:"redis_url"`
	Namespace string `yaml:"namespace"`
}

// PoolConfig keeps worker pool configurations.
type PoolConfig struct {
	// Worker concurrency
	WorkerCount  uint             `yaml:"workers"`
	Backend      string           `yaml:"backend"`
	RedisPoolCfg *RedisPoolConfig `yaml:"redis_pool,omitempty"`
}

// LoggerConfig keeps logger configurations.
type LoggerConfig struct {
	BasePath      string `yaml:"path"`
	LogLevel      string `yaml:"level"`
	ArchivePeriod uint   `yaml:"archive_period"`
}

// Load the configuration options from the specified yaml file.
// If the yaml file is specified and existing, load configurations from yaml file first;
// If detecting env variables is specified, load configurations from env variables;
// Please pay attentions, the detected env variable will override the same configuration item loading from file.
//
// yamlFilePath string: The path config yaml file
// readEnv       bool  : Whether detect the environment variables or not
func (c *Configuration) Load(yamlFilePath string, detectEnv bool) error {
	if !utils.IsEmptyStr(yamlFilePath) {
		// Try to load from file first
		data, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return err
		}

		if err = yaml.Unmarshal(data, c); err != nil {
			return err
		}
	}

	if detectEnv {
		// Load from env variables
		c.loadEnvs()
	}

	// translate redis url if needed
	if c.PoolConfig != nil && c.PoolConfig.RedisPoolCfg != nil {
		redisAddress := c.PoolConfig.RedisPoolCfg.RedisURL

		if !utils.IsEmptyStr(redisAddress) {
			if _, err := url.Parse(redisAddress); err != nil {
				if redisURL, ok := utils.TranslateRedisAddress(redisAddress); ok {
					c.PoolConfig.RedisPoolCfg.RedisURL = redisURL
				}
			} else {
				if !strings.HasPrefix(redisAddress, redisSchema) {
					c.PoolConfig.RedisPoolCfg.RedisURL = fmt.Sprintf("%s%s", redisSchema, redisAddress)
				}
			}
		}
	}

	// Validate settings
	return c.validate()
}

// GetLogBasePath returns the log base path config
func GetLogBasePath() string {
	if DefaultConfig.LoggerConfig != nil {
		return DefaultConfig.LoggerConfig.BasePath
	}

	return ""
}

// GetLogLevel returns the log level
func GetLogLevel() string {
	if DefaultConfig.LoggerConfig != nil {
		return DefaultConfig.LoggerConfig.LogLevel
	}

	return ""
}

// GetLogArchivePeriod returns the archive period
func GetLogArchivePeriod() uint {
	if DefaultConfig.LoggerConfig != nil {
		return DefaultConfig.LoggerConfig.ArchivePeriod
	}

	return 1 // return default
}

// GetUIAuthSecret get the auth secret of UI side
func GetUIAuthSecret() string {
	return utils.ReadEnv(uiAuthSecret)
}

// Load env variables
func (c *Configuration) loadEnvs() {
	prot := utils.ReadEnv(jobServiceProtocol)
	if !utils.IsEmptyStr(prot) {
		c.Protocol = prot
	}

	p := utils.ReadEnv(jobServicePort)
	if !utils.IsEmptyStr(p) {
		if po, err := strconv.Atoi(p); err == nil {
			c.Port = uint(po)
		}
	}

	// Only when protocol is https
	if c.Protocol == JobServiceProtocolHTTPS {
		cert := utils.ReadEnv(jobServiceHTTPCert)
		if !utils.IsEmptyStr(cert) {
			if c.HTTPSConfig != nil {
				c.HTTPSConfig.Cert = cert
			} else {
				c.HTTPSConfig = &HTTPSConfig{
					Cert: cert,
				}
			}
		}

		certKey := utils.ReadEnv(jobServiceHTTPKey)
		if !utils.IsEmptyStr(certKey) {
			if c.HTTPSConfig != nil {
				c.HTTPSConfig.Key = certKey
			} else {
				c.HTTPSConfig = &HTTPSConfig{
					Key: certKey,
				}
			}
		}
	}

	backend := utils.ReadEnv(jobServiceWorkerPoolBackend)
	if !utils.IsEmptyStr(backend) {
		if c.PoolConfig == nil {
			c.PoolConfig = &PoolConfig{}
		}
		c.PoolConfig.Backend = backend
	}

	workers := utils.ReadEnv(jobServiceWorkers)
	if !utils.IsEmptyStr(workers) {
		if count, err := strconv.Atoi(workers); err == nil {
			if c.PoolConfig == nil {
				c.PoolConfig = &PoolConfig{}
			}
			c.PoolConfig.WorkerCount = uint(count)
		}
	}

	if c.PoolConfig != nil && c.PoolConfig.Backend == JobServicePoolBackendRedis {
		redisURL := utils.ReadEnv(jobServiceRedisURL)
		if !utils.IsEmptyStr(redisURL) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			c.PoolConfig.RedisPoolCfg.RedisURL = redisURL
		}

		rn := utils.ReadEnv(jobServiceRedisNamespace)
		if !utils.IsEmptyStr(rn) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			c.PoolConfig.RedisPoolCfg.Namespace = rn
		}
	}

	// logger
	loggerPath := utils.ReadEnv(jobServiceLoggerBasePath)
	if !utils.IsEmptyStr(loggerPath) {
		if c.LoggerConfig == nil {
			c.LoggerConfig = &LoggerConfig{}
		}
		c.LoggerConfig.BasePath = loggerPath
	}

	loggerLevel := utils.ReadEnv(jobServiceLoggerLevel)
	if !utils.IsEmptyStr(loggerLevel) {
		if c.LoggerConfig == nil {
			c.LoggerConfig = &LoggerConfig{}
		}
		c.LoggerConfig.LogLevel = loggerLevel
	}

	archivePeriod := utils.ReadEnv(jobServiceLoggerArchivePeriod)
	if !utils.IsEmptyStr(archivePeriod) {
		if period, err := strconv.Atoi(archivePeriod); err == nil {
			if c.LoggerConfig == nil {
				c.LoggerConfig = &LoggerConfig{}
			}
			c.LoggerConfig.ArchivePeriod = uint(period)
		}
	}
}

// Check if the configurations are valid settings.
func (c *Configuration) validate() error {
	if c.Protocol != JobServiceProtocolHTTPS &&
		c.Protocol != JobServiceProtocolHTTP {
		return fmt.Errorf("protocol should be %s or %s, but current setting is %s",
			JobServiceProtocolHTTP,
			JobServiceProtocolHTTPS,
			c.Protocol)
	}

	if !utils.IsValidPort(c.Port) {
		return fmt.Errorf("port number should be a none zero integer and less or equal 65535, but current is %d", c.Port)
	}

	if c.Protocol == JobServiceProtocolHTTPS {
		if c.HTTPSConfig == nil {
			return fmt.Errorf("certificate must be configured if serve with protocol %s", c.Protocol)
		}

		if utils.IsEmptyStr(c.HTTPSConfig.Cert) ||
			!utils.FileExists(c.HTTPSConfig.Cert) ||
			utils.IsEmptyStr(c.HTTPSConfig.Key) ||
			!utils.FileExists(c.HTTPSConfig.Key) {
			return fmt.Errorf("certificate for protocol %s is not correctly configured", c.Protocol)
		}
	}

	if c.PoolConfig == nil {
		return errors.New("no worker pool is configured")
	}

	if c.PoolConfig.Backend != JobServicePoolBackendRedis {
		return fmt.Errorf("worker pool backend %s does not support", c.PoolConfig.Backend)
	}

	// When backend is redis
	if c.PoolConfig.Backend == JobServicePoolBackendRedis {
		if c.PoolConfig.RedisPoolCfg == nil {
			return fmt.Errorf("redis pool must be configured when backend is set to '%s'", c.PoolConfig.Backend)
		}
		if utils.IsEmptyStr(c.PoolConfig.RedisPoolCfg.RedisURL) {
			return errors.New("URL of redis pool is empty")
		}

		if !strings.HasPrefix(c.PoolConfig.RedisPoolCfg.RedisURL, redisSchema) {
			return errors.New("Invalid redis URL")
		}

		if _, err := url.Parse(c.PoolConfig.RedisPoolCfg.RedisURL); err != nil {
			return fmt.Errorf("Invalid redis URL: %s", err.Error())
		}

		if utils.IsEmptyStr(c.PoolConfig.RedisPoolCfg.Namespace) {
			return errors.New("namespace of redis pool is required")
		}
	}

	if c.LoggerConfig == nil {
		return errors.New("missing logger config")
	}

	if !utils.DirExists(c.LoggerConfig.BasePath) {
		return fmt.Errorf("logger path %s not exists or should be dir", c.LoggerConfig.BasePath)
	}

	validLevels := "DEBUG,INFO,WARNING,ERROR,FATAL"
	if !strings.Contains(validLevels, c.LoggerConfig.LogLevel) {
		return fmt.Errorf("logger level can only be one of: %s", validLevels)
	}

	if c.LoggerConfig.ArchivePeriod == 0 {
		return fmt.Errorf("logger archive period should be greater than 0")
	}

	return nil // valid
}
