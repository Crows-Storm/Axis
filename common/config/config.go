package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const insecureDefaultJWTSecret = "default-jwt-secret-change-in-production"

const minJWTSecretLength = 32

var global *Config

func Get() *Config {
	return global
}

// GetServerAddr build Rest server addr
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.RestPort)
}

// GetGRPCAddr build GRPC server addr
func (c *Config) GetGRPCAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.GRPCPort)
}

// Config is .env all config
type Config struct {
	ServerName               string
	ServerHost               string
	LogLevel                 string
	RestPort                 int
	JWTSecret                string
	GRPCPort                 int
	ServiceDiscoveryConfig   ConsulConfig
	DbConfig                 DBConfig
	RedisHealthCheckInterval int
	ReadRedis                ReadRedisConfig
	WriteRedis               WriteRedisConfig
}

type ConsulConfig struct {
	Host     string
	Port     int
	ACTToken string
	Timeout  time.Duration
}
type DBConfig struct {
	DBType     string
	DBPath     string
	DBHost     string
	DBPort     int
	DBLoginId  string
	DBPassword string
	DBSchema   string
	DBSslMode  string
}

type ReadRedisConfig struct {
	Address      []string
	DB           int
	DisableCache bool
	Password     string
}

type WriteRedisConfig struct {
	Address      []string
	DB           int
	DisableCache bool
	Password     string
}

func MustInit() {
	if err := initConfig(); err != nil {
		panic(fmt.Sprintf("[CONFIG MustInit] Init config ERROR: %v\n", err))
	}
}

// Init initializes global configuration (from .env). Prefer MustInit from main.
func Init() {
	if err := initConfig(); err != nil {
		// Preserve historical fail-soft behavior for non-main callers (tests, tools);
		// the process can still observe the error via Get() returning nil.
		fmt.Fprintf(os.Stderr, "[CONFIG MustInit] config init failed: %v\n", err)
	}
}

// initConfig initializes all components based on the .env
func initConfig() error {
	cfg := &Config{
		ServerName:               getEnvAsString("SERVER_NAME", ""),
		ServerHost:               getEnvAsString("SERVER_HOST", "localhost"),
		LogLevel:                 getEnvAsString("LOG_LEVEL", "info"),
		RestPort:                 getEnvAsInt("REST_PORT", 0),
		JWTSecret:                getEnvAsString("JWT_SECRET", insecureDefaultJWTSecret),
		ServiceDiscoveryConfig:   getServiceDiscoveryConfig(),
		GRPCPort:                 getEnvAsInt("GRPC_PORT", 0),
		DbConfig:                 getDBConfig(),
		RedisHealthCheckInterval: getRedisHealthCheckInterval(),
		ReadRedis:                getReadRedis(),
		WriteRedis:               getWriteRedis(),
	}
	global = cfg

	return nil
}

func getServiceDiscoveryConfig() ConsulConfig {
	return ConsulConfig{
		Host:     getEnvAsString("CONSUL_HOST", "localhost"),
		Port:     getEnvAsInt("CONSUL_PORT", 8500),
		ACTToken: getEnvAsString("CONSUL_ACL_TOKEN", ""),
		Timeout:  time.Duration(getEnvAsInt("CONSUL_TIMEOUT", 5)) * time.Second,
	}
}

func getRedisHealthCheckInterval() int {
	interval := os.Getenv("REDIS_HEALTH_CHECK_INTERVAL")
	if interval == "" {
		return 30
	}
	val, err := strconv.Atoi(interval)
	if err != nil {
		panic(fmt.Sprintf("REDIS_HEALTH_CHECK_INTERVAL must be a valid integer, got: %s", interval))
	}
	return val
}

func getDBConfig() DBConfig {
	return DBConfig{
		DBType:     getEnvAsString("DB_TYPE", ""),
		DBPath:     getEnvAsString("DB_PATH", ""),
		DBHost:     getEnvAsString("DB_HOST", ""),
		DBPort:     getEnvAsInt("DB_PORT", 3306),
		DBLoginId:  getEnvAsString("DB_LOGIN_ID", ""),
		DBPassword: getEnvAsString("DB_PASSWORD", ""),
		DBSchema:   getEnvAsString("DB_SCHEMA", ""),
		DBSslMode:  getEnvAsString("DB_SSL_MODE", "disable"),
	}
}

func getEnvAsString(key string, def string) string {
	value := os.Getenv(key)
	if value == "" && def == "" {
		panic(fmt.Sprintf("The .env file is missing the %s configuration.", key))
	} else if def != "" {
		value = def
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("Environment variable %s must be a valid integer, got: %s", key, value))
	}
	return val
}

// getEnvAsBool Retrieves the environment variable and converts it to a boolean; if it does not exist, it returns the default value.
func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Sprintf("Environment variable %s must be a valid boolean, got: %s", key, value))
	}
	return val
}

func getReadRedis() ReadRedisConfig {
	addressStr := getEnvAsString("WRITE_REDIS_ADDRESS", "localhost:6379")
	addresses := parseRedisAddresses(addressStr)
	return ReadRedisConfig{
		Address:      addresses,
		DB:           getEnvAsInt("READ_REDIS_DB", 0),
		DisableCache: getEnvAsBool("READ_REDIS_DISABLE_CACHE", false),
		Password:     getEnvAsString("READ_REDIS_PASSWORD", ""),
	}
}

func getWriteRedis() WriteRedisConfig {
	addressStr := getEnvAsString("READ_REDIS_ADDRESS", "localhost:6379")
	addresses := parseRedisAddresses(addressStr)
	return WriteRedisConfig{
		Address:      addresses,
		DB:           getEnvAsInt("WRITE_REDIS_DB", 0),
		DisableCache: getEnvAsBool("WRITE_REDIS_DISABLE_CACHE", false),
		Password:     getEnvAsString("WRITE_REDIS_PASSWORD", ""),
	}
}

func parseRedisAddresses(addressStr string) []string {
	if addressStr == "" {
		return []string{"localhost:6379"}
	}

	parts := strings.Split(addressStr, ",")
	var addresses []string
	for _, addr := range parts {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			addresses = append(addresses, addr)
		}
	}

	if len(addresses) == 0 {
		return []string{"localhost:6379"}
	}

	return addresses
}
