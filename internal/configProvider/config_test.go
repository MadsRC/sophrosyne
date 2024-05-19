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

//go:build !integration

package configProvider

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/madsrc/sophrosyne/internal/validator"

	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
)

var testYamlFilePath = "testdata/config.yaml"
var securitySaltFilePath = "testdata/security.salt"
var securitySiteKeyFilePath = "testdata/security.sitekey"
var testFilePath = "testdata/test.file"
var databaseUserKey = "database.user"
var databasePasswordKey = "database.password"
var nonExistentYamlFilePath = "testdata/non-existent.yaml"
var newPasswordString = "new-password 42 c@t"
var rootConfigYamlPath = "/config.yaml"

func TestSecretFromFile(t *testing.T) {
	cases := []struct {
		name    string
		file    string
		keyName string
	}{
		{
			name:    "secret file with text",
			file:    testFilePath,
			keyName: "test.file",
		},
		{
			name:    "secret file with binary data",
			file:    "testdata/file.binary",
			keyName: "file.binary",
		},
		{
			name:    "secret file with empty data",
			file:    "testdata/empty.file",
			keyName: "empty.file",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the raw file for comparison
			dat, err := os.ReadFile(tc.file)
			require.NoError(t, err)

			// Read the secret from the file
			secret, err := secretFromFile(tc.file)
			require.NoError(t, err)

			// Compare the raw file data with the secret
			require.Equal(t, dat, secret[tc.keyName])
		})
	}

}

func TestSecretFromFileErrors(t *testing.T) {
	cases := []struct {
		name string
		file string
	}{
		{
			name: "empty filepath",
			file: "",
		},
		{
			name: "non-existent file",
			file: "testdata/non-existent.file",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the secret from the file
			secret, err := secretFromFile(tc.file)
			require.Error(t, err)
			require.Nil(t, secret)
		})
	}

}

func TestEnvExtractor(t *testing.T) {
	cases := []struct {
		name string
		key  string
		val  string
		exp  interface{}
	}{
		{
			name: "single value",
			key:  "KEY",
			val:  "value",
			exp:  "value",
		},
		{
			name: "multiple values",
			key:  "KEY",
			val:  "value1 value2 value3",
			exp:  []string{"value1", "value2", "value3"},
		},
		{
			name: "empty value",
			key:  "KEY",
			val:  "",
			exp:  "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, val := envExtractor(tc.key, tc.val)
			require.Equal(t, strings.ToLower(tc.key), key)
			require.Equal(t, tc.exp, val)
		})
	}
}

func TestLoadYamlConfig(t *testing.T) {
	k := koanf.New(sophrosyne.ConfigDelimiter)
	err := loadYamlConfig(k, file.Provider(testYamlFilePath))
	require.NoError(t, err)

	require.Equal(t, "postgres", k.String(databaseUserKey))
	require.Equal(t, "postgres", k.String(databasePasswordKey))
	require.Equal(t, "localhost", k.String("database.host"))
	require.Equal(t, "5432", k.String("database.port"))
	require.Equal(t, "postgres", k.String("database.name"))
}

func TestLoadYamlConfigErrors(t *testing.T) {
	k := koanf.New(sophrosyne.ConfigDelimiter)
	err := loadYamlConfig(k, file.Provider(nonExistentYamlFilePath))
	require.Error(t, err)
}

func TestLoadConfig(t *testing.T) {
	defaultConfig := map[string]interface{}{
		"test.file": "not content",
	}
	k := koanf.New(sophrosyne.ConfigDelimiter)
	err := loadConfig(
		k,
		defaultConfig,
		file.Provider(testYamlFilePath),
		nil,
		[]string{testFilePath},
	)
	require.NoError(t, err)

	require.Equal(t, "postgres", k.String(databaseUserKey))
	require.Equal(t, "postgres", k.String(databasePasswordKey))
	require.Equal(t, "localhost", k.String("database.host"))
	require.Equal(t, "5432", k.String("database.port"))
	require.Equal(t, "postgres", k.String("database.name"))
	require.NotEqual(t, "content", k.String("not content"))
}

func TestLoadConfigErr(t *testing.T) {
	cases := []struct {
		name        string
		defaultConf map[string]interface{}
		yamlFile    string
		secretFiles []string
	}{
		{
			name:        "non-existent yaml file",
			defaultConf: map[string]interface{}{},
			yamlFile:    nonExistentYamlFilePath,
			secretFiles: []string{},
		},
		{
			name:        "non-existent secret file",
			defaultConf: map[string]interface{}{},
			yamlFile:    testYamlFilePath,
			secretFiles: []string{"testdata/non-existent.file"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k := koanf.New(sophrosyne.ConfigDelimiter)
			err := loadConfig(
				k,
				tc.defaultConf,
				file.Provider(tc.yamlFile),
				nil,
				tc.secretFiles,
			)
			require.Error(t, err)
		})
	}
}

func TestNewConfigProvider(t *testing.T) {
	initialPw := "password"
	yamlContent := []byte(`database:
  password: ` + initialPw)

	tempDir := t.TempDir()
	tempFile := tempDir + rootConfigYamlPath
	err := os.WriteFile(tempFile, yamlContent, 0644)
	require.NoError(t, err)

	c, err := NewConfigProvider(tempFile, nil, []string{securitySaltFilePath, securitySiteKeyFilePath}, validator.NewValidator())
	require.NoError(t, err)

	require.Equal(t, initialPw, c.k.String(databasePasswordKey))

	newYamlContent := []byte(`database:
  password: ` + newPasswordString)
	err = os.WriteFile(tempFile, newYamlContent, 0644)
	require.NoError(t, err)

	// sleep for 100 milliseconds to allow the file watcher to pick up the
	// change and reload the config
	time.Sleep(100 * time.Millisecond)

	// Ensure that the config has been updated - Allow for one failure to account
	// for potentially slow CI runners.
	if !assert.Equal(t, newPasswordString, c.k.String(databasePasswordKey)) {
		time.Sleep(400 * time.Millisecond)
		require.Equal(t, newPasswordString, c.k.String(databasePasswordKey))
	}

	badYamlContent := []byte{0x00, 0x01, 0x02}
	err = os.WriteFile(tempFile, badYamlContent, 0644)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Ensure that the config has been updated - Allow for one failure to account
	// for potentially slow CI runners.
	// The bad yaml content should not have been loaded and thus the previous
	// value should still be present.
	if !assert.Equal(t, newPasswordString, c.k.String(databasePasswordKey)) {
		time.Sleep(400 * time.Millisecond)
		require.Equal(t, newPasswordString, c.k.String(databasePasswordKey))
	}
}

func TestNewConfigProviderErrorNoYamlFile(t *testing.T) {
	c, err := NewConfigProvider(nonExistentYamlFilePath, nil, []string{testFilePath}, nil)
	require.Error(t, err)
	require.Nil(t, c)
}

func TestNewConfigProviderErrorUpdateFailValidate(t *testing.T) {
	tempDir := t.TempDir()
	// Copy file at testYamlFilePath to tempDir
	tempFile := tempDir + rootConfigYamlPath
	yamlContent, err := os.ReadFile(testYamlFilePath)
	require.NoError(t, err)
	err = os.WriteFile(tempFile, yamlContent, 0644)
	require.NoError(t, err)

	c, err := NewConfigProvider(tempFile, nil, []string{securitySaltFilePath, securitySiteKeyFilePath}, validator.NewValidator())
	require.NoError(t, err)

	cfg := c.Get()
	require.NotNil(t, cfg)
	require.Equal(t, "postgres", cfg.Database.User)
	require.Equal(t, "postgres", cfg.Database.Password)
	require.Equal(t, "localhost", cfg.Database.Host)
	require.Equal(t, 5432, cfg.Database.Port)
	require.Equal(t, "postgres", cfg.Database.Name)

	newYamlContent := []byte(`database:
  port: 65536`)
	err = os.WriteFile(tempFile, newYamlContent, 0644)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	// Ensure that the config has been updated - Allow for one failure to account
	// for potentially slow CI runners.
	if !assert.Equal(t, 5432, cfg.Database.Port) {
		time.Sleep(400 * time.Millisecond)
		require.Equal(t, 5432, cfg.Database.Port)
	}
}

func TestNewConfigProviderErrorValidateYamlFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := tempDir + rootConfigYamlPath
	yamlContent := []byte(`database:
  port: not-a-number`)
	err := os.WriteFile(tempFile, yamlContent, 0644)
	require.NoError(t, err)

	c, err := NewConfigProvider(tempFile, nil, []string{testFilePath}, validator.NewValidator())
	require.Error(t, err)
	require.Nil(t, c)
}

func TestConfigProviderGet(t *testing.T) {
	tempDir := t.TempDir()
	// Copy file at testYamlFilePath to tempDir
	tempFile := tempDir + rootConfigYamlPath
	yamlContent, err := os.ReadFile(testYamlFilePath)
	require.NoError(t, err)
	err = os.WriteFile(tempFile, yamlContent, 0644)
	require.NoError(t, err)

	c, err := NewConfigProvider(tempFile, nil, []string{securitySaltFilePath, securitySiteKeyFilePath}, validator.NewValidator())
	require.NoError(t, err)

	cfg := c.Get()
	require.NotNil(t, cfg)
	require.Equal(t, "postgres", cfg.Database.User)
	require.Equal(t, "postgres", cfg.Database.Password)
	require.Equal(t, "localhost", cfg.Database.Host)
	require.Equal(t, 5432, cfg.Database.Port)
	require.Equal(t, "postgres", cfg.Database.Name)

	newYamlContent := []byte(`database:
  password: ` + newPasswordString)
	err = os.WriteFile(tempFile, newYamlContent, 0644)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	require.NotNil(t, cfg)
	// Ensure that the config has been updated - Allow for one failure to account
	// for potentially slow CI runners.
	if !assert.Equal(t, newPasswordString, cfg.Database.Password) {
		time.Sleep(900 * time.Millisecond)
		require.Equal(t, newPasswordString, cfg.Database.Password)
	}
	// Ensure that the other values have not changed
	require.Equal(t, "postgres", cfg.Database.User)
	require.Equal(t, "localhost", cfg.Database.Host)
	require.Equal(t, 5432, cfg.Database.Port)
	require.Equal(t, "postgres", cfg.Database.Name)

	badYamlContent := []byte{0x00, 0x01, 0x02}
	err = os.WriteFile(tempFile, badYamlContent, 0644)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// The bad yaml content should not have been loaded and thus the previous
	// value should still be present.
	require.NotNil(t, cfg)
	// Ensure that the config has been updated - Allow for one failure to account
	// for potentially slow CI runners.
	if !assert.Equal(t, newPasswordString, cfg.Database.Password) {
		time.Sleep(400 * time.Millisecond)
		require.Equal(t, newPasswordString, cfg.Database.Password)
	}
	require.Equal(t, "postgres", cfg.Database.User)
	require.Equal(t, "localhost", cfg.Database.Host)
	require.Equal(t, 5432, cfg.Database.Port)
	require.Equal(t, "postgres", cfg.Database.Name)
}
