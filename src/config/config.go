package config

import (
	"os"
	"strconv"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
)

// ConfigService provides typed, safe access to environment variables.
// It is designed to be injected into other services via the IoC container.
//
// Register it in your module using ConfigModule, or add NewConfigService
// directly to a module's Providers list.
type ConfigService struct{}

// NewConfigService creates a new ConfigService.
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// Get returns the value of the environment variable named by key.
// Returns an empty string if the variable is not set.
func (c *ConfigService) Get(key string) string {
	return os.Getenv(key)
}

// GetOrDefault returns the value of key, or fallback if key is not set or empty.
func (c *ConfigService) GetOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetInt parses key as a base-10 integer.
// Returns fallback if the variable is not set or cannot be parsed.
func (c *ConfigService) GetInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// GetBool parses key as a boolean.
// Accepts "1", "true", "yes" as true; "0", "false", "no" as false.
// Returns fallback for any other value or unset variable.
func (c *ConfigService) GetBool(key string, fallback bool) bool {
	switch os.Getenv(key) {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	default:
		return fallback
	}
}

// MustGet returns the value of key.
// Panics if the variable is not set or empty, ensuring the application
// fails fast on missing required configuration.
//
//	dsn := cfg.MustGet("DATABASE_URL")
func (c *ConfigService) MustGet(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("nexgou/config: required environment variable not set: " + key)
	}
	return v
}

// ── ConfigModule ──────────────────────────────────────────────────────────────

// ConfigModule is a ready-to-use Nexgou module that registers and exports
// ConfigService. Import it in any module that needs environment variable access.
//
// Usage:
//
//	var AppModule = nexgou.Module(nexgou.ModuleOptions{
//	    Imports: []nexgou.IModule{nexgou.ConfigModule, UserModule},
//	})
//
//	func NewUserService(cfg *config.ConfigService) *UserService {
//	    port := cfg.GetInt("PORT", 3000)
//	    dsn  := cfg.MustGet("DATABASE_URL")
//	}
var ConfigModule common.IModule = core.NewModule(common.ModuleOptions{
	Providers: []any{NewConfigService},
	Exports:   []any{NewConfigService},
})
