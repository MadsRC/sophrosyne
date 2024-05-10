package configProvider

import (
	"encoding/hex"
	"path/filepath"
	"strings"
	"sync"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/madsrc/sophrosyne"
)

func envExtractor(s string, v string) (string, interface{}) {
	key := strings.Replace(strings.ToLower(strings.TrimPrefix(s, sophrosyne.ConfigEnvironmentPrefix)), "_", sophrosyne.ConfigDelimiter, -1)

	if strings.Contains(v, " ") {
		return key, strings.Split(v, " ")
	}

	if strings.HasPrefix(v, "0x") {
		b, err := hex.DecodeString(v[2:])
		if err == nil {
			return key, b
		}
	}

	return key, v
}

func loadConfig(k *koanf.Koanf, defaultConfig map[string]interface{}, yamlFile koanf.Provider, overwrites map[string]interface{}, secretFiles []string) error {
	k.Load(confmap.Provider(defaultConfig, sophrosyne.ConfigDelimiter), nil)

	if err := loadYamlConfig(k, yamlFile); err != nil {
		return err
	}

	k.Load(env.ProviderWithValue(sophrosyne.ConfigEnvironmentPrefix, sophrosyne.ConfigDelimiter, envExtractor), nil)

	k.Load(confmap.Provider(overwrites, sophrosyne.ConfigDelimiter), nil)

	for _, secretFile := range secretFiles {
		secret, err := secretFromFile(secretFile)
		if err != nil {
			return err
		}
		k.Load(confmap.Provider(secret, sophrosyne.ConfigDelimiter), nil)
	}

	return nil
}

func loadYamlConfig(k *koanf.Koanf, yamlFile koanf.Provider) error {
	if err := k.Load(yamlFile, yaml.Parser()); err != nil {
		return err
	}
	return nil
}

type ConfigProvider struct {
	k        *koanf.Koanf
	config   *sophrosyne.Config
	validate sophrosyne.Validator
	mu       sync.Mutex
}

func NewConfigProvider(yamlFilePath string, overwrites map[string]interface{}, secretFiles []string, validator sophrosyne.Validator) (*ConfigProvider, error) {
	cfgProv := &ConfigProvider{
		config:   &sophrosyne.Config{},
		k:        koanf.New(sophrosyne.ConfigDelimiter),
		validate: validator,
	}

	yamlFile := file.Provider(yamlFilePath)

	if err := loadConfig(cfgProv.k, sophrosyne.DefaultConfig, yamlFile, overwrites, secretFiles); err != nil {
		return nil, err
	}

	yamlFile.Watch(func(event interface{}, err error) {
		if err != nil {
			// Error occurred when watching the file.
			return
		}
		// We have to reload not just the yaml file, but everything else as well.
		// If we do not, we risk that values that have been removed from the
		// yaml file are still present in the config.
		err = loadConfig(cfgProv.k, sophrosyne.DefaultConfig, yamlFile, overwrites, secretFiles)
		if err != nil {
			// Error occurred when reloading the yaml file.
			return
		}
		newConf := &sophrosyne.Config{}
		cfgProv.k.UnmarshalWithConf("", newConf, koanf.UnmarshalConf{Tag: "key"})
		err = cfgProv.validate.Validate(newConf)
		if err != nil {
			// Error occurred when validating the config.
			return
		}
		// Reuse the existing pointer (as this is what the user is already
		// using) and just copy the new values over.
		cfgProv.mu.Lock()
		defer cfgProv.mu.Unlock()
		*cfgProv.config = *newConf
	})

	cfgProv.k.UnmarshalWithConf("", cfgProv.config, koanf.UnmarshalConf{Tag: "key"})

	err := cfgProv.validate.Validate(cfgProv.config)
	if err != nil {
		return nil, err
	}

	return cfgProv, nil
}

func (c *ConfigProvider) Get() *sophrosyne.Config {
	return c.config
}

func secretFromFile(filePath string) (map[string]interface{}, error) {
	fileName := filepath.Base(filePath)

	f := file.Provider(filePath)

	b, err := f.ReadBytes()

	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	out[fileName] = b

	return out, nil
}
