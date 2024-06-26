// Sophrosyne
//   Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Affero General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Affero General Public License for more details.
//
//   You should have received a copy of the GNU Affero General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.

package sophrosyne

import "time"

// The ConfigProvider interface is used to retrieve the configuration of the
// application.
//
// Implementations may support reloading the configuration by watching
// configuration sources for changes. In the event that the configuration is
// reloaded, the implementation must ensure that the pointer address
// returned by the Get method remains the same, but is expected to change the
// object pointed to by the pointer.
//
// Additionally, implementations should ensure that the configuration is
//
//	based off of the DefaultConfig and validated using the validate
//
// information in the Config struct's validate tag.
//
// The ConfigProvider interface is expected to be thread-safe.
//
// The ConfigProvider interface is expected to be used as a singleton.
//
// The ConfigProvider interface is expected to reference the
// [ConfigEnvironmentPrefix] if reading from the environment.
//
// The ConfigProvider interface is expected to use the [ConfigDelimiter] to
// separate keys in the configuration.
//
// The Get method returns the configuration of the application. Multiple calls
// to Get must return same pointer address.
type ConfigProvider interface {
	Get() *Config
}

// Default configuration for the application. ConfigProvider implementations
// should use this configuration as the default configuration.
//
// Values that should not have a default value should not be included.
var DefaultConfig = map[string]interface{}{
	"database.user":                           "postgres",
	"database.host":                           "localhost",
	"database.port":                           5432,
	"database.name":                           "postgres",
	"server.port":                             8080,
	"logging.level":                           LogLevelInfo,
	"logging.format":                          LogFormatJSON,
	"logging.enabled":                         true,
	"tracing.enabled":                         true,
	"tracing.batch.timeout":                   5,
	"tracing.output":                          OtelOutputStdout,
	"metrics.enabled":                         false,
	"metrics.interval":                        60,
	"metrics.output":                          OtelOutputStdout,
	"principals.root.name":                    "root",
	"principals.root.email":                   "root@localhost",
	"principals.root.recreate":                false,
	"services.users.pageSize":                 2,
	"services.users.cache.TTL":                1 * time.Second,
	"services.users.cache.cleanupInterval":    500 * time.Millisecond,
	"security.tls.keyType":                    "EC-P384",
	"security.tls.insecureSkipVerify":         false,
	"services.profiles.pageSize":              2,
	"services.profiles.cache.TTL":             1 * time.Second,
	"services.profiles.cache.cleanupInterval": 500 * time.Millisecond,
	"services.checks.pageSize":                2,
	"services.checks.cache.TTL":               1 * time.Second,
	"services.checks.cache.cleanupInterval":   500 * time.Millisecond,
	"server.maxBodySize":                      20 * megabyte,
	"server.advertisedHost":                   "localhost",
}

const megabyte int64 = 1048576

// The Config struct is used to store the configuration of the application.
//
// The ConfigProvider interface is used to retrieve the configuration of the
// application from the environment variables, configuration files, and secret
// files.
//
// The validate tag is used to validate the configuration using
// https://github.com/go-playground/validator/v10.
type Config struct {
	Principals struct {
		Root struct {
			Name     string `key:"name" validate:"required"`
			Email    string `key:"email" validate:"required"`
			Recreate bool   `key:"recreate"`
		} `key:"root" validate:"required"`
	} `key:"principals" validate:"required"`
	Database struct {
		User     string `key:"user" validate:"required"`
		Password string `key:"password" validate:"required"`
		Host     string `key:"host" validate:"required"`
		Port     int    `key:"port" validate:"required,min=1,max=65535"`
		Name     string `key:"name" validate:"required"`
	} `key:"database"`
	Server  ServerConfig `key:"server"`
	Logging struct {
		Enabled bool      `key:"enabled"`
		Level   LogLevel  `key:"level" validate:"required,oneof=debug info"`
		Format  LogFormat `key:"format" validate:"required,oneof=text json"`
	} `key:"logging"`
	Tracing struct {
		Enabled bool `key:"enabled"`
		Batch   struct {
			Timeout int `key:"timeout"`
		} `key:"batch"`
		Output OtelOutput `key:"output" validate:"required,oneof=stdout http"`
	} `key:"tracing"`
	Metrics struct {
		Enabled  bool       `key:"enabled"`
		Interval int        `key:"interval"`
		Output   OtelOutput `key:"output" validate:"required,oneof=stdout http"`
	} `key:"metrics"`
	Security SecurityConfig `key:"security" validate:"required"`
	Services struct {
		Users struct {
			PageSize int         `key:"pageSize" validate:"required,min=2"`
			Cache    CacheConfig `key:"cache" validate:"required"`
		} `key:"users" validate:"required"`
		Profiles struct {
			PageSize int         `key:"pageSize" validate:"required,min=2"`
			Cache    CacheConfig `key:"cache" validate:"required"`
		} `key:"profiles" validate:"required"`
		Checks struct {
			PageSize int         `key:"pageSize" validate:"required,min=2"`
			Cache    CacheConfig `key:"cache" validate:"required"`
		} `key:"checks" validate:"required"`
	} `key:"services" validate:"required"`
	Development struct {
		StaticRootToken string `key:"staticRootToken"`
	} `key:"development"`
}

type CacheConfig struct {
	TTL             time.Duration `key:"ttl" validate:"required,min=1"`
	CleanupInterval time.Duration `key:"cleanupInterval" validate:"required,min=1"`
}

type TLSConfig struct {
	KeyType            string `key:"keyType" validate:"required,oneof=RSA-4096 EC-P224 EC-P256 EC-P384 EC-P521 ED25519"`
	CertificatePath    string `key:"certificatePath"`
	KeyPath            string `key:"keyPath"`
	InsecureSkipVerify bool   `key:"insecureSkipVerify"`
}

type SecurityConfig struct {
	SiteKey []byte    `key:"siteKey" validate:"required,min=64,max=64"`
	Salt    []byte    `key:"salt" validate:"required,min=32,max=32"`
	TLS     TLSConfig `key:"tls" validate:"required"`
}

type ServerConfig struct {
	Port           int    `key:"port" validate:"required,min=1,max=65535"`
	MaxBodySize    int64  `key:"maxBodySize" validate:"required,min=1"` // in bytes
	AdvertisedHost string `key:"advertisedHost" validate:"required"`
}

// ConfigEnvironmentPrefix is the prefix used to identify the environment
// variables that are used to configure the application.
var ConfigEnvironmentPrefix = "SOPH_"

// ConfigDelimiter is the delimiter used to separate the keys in the
// configuration.
var ConfigDelimiter = "."
