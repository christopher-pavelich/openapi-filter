package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var ErrConfigPathEmpty = errors.New("config path is empty")

func initConfig[C any](configPath string) (*C, error) {
	k := koanf.New(".")

	configExt := strings.TrimLeft(filepath.Ext(configPath), ".")

	var parser koanf.Parser
	switch configExt {
	case "yaml", "yml":
		parser = yaml.Parser()
	case "toml":
		parser = toml.Parser()
	case "json":
		parser = json.Parser()
	default:
		return nil, fmt.Errorf("unsupported config format: %s", configExt)
	}

	if err := k.Load(file.Provider(configPath), parser); err != nil {
		return nil, fmt.Errorf("k.Load: %w", err)
	}

	var cfg C
	// Use koanf's Unmarshal with custom mapstructure hook
	unmarshalOpts := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:           &cfg,
			DecodeHook:       pathConfigDecodeHook,
			WeaklyTypedInput: true,
		},
	}
	if err := k.UnmarshalWithConf("", &cfg, unmarshalOpts); err != nil {
		return nil, fmt.Errorf("k.UnmarshalWithConf: %w", err)
	}
	return &cfg, nil
}

// pathConfigDecodeHook is a mapstructure decode hook that handles PathConfig decoding
// from both simple array format and advanced object format.
func pathConfigDecodeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	// Handle conversion to PathConfig (both value and pointer types)
	pathConfigType := reflect.TypeOf(PathConfig{})
	pathConfigPtrType := reflect.TypeOf((*PathConfig)(nil))

	if to != pathConfigType && to != pathConfigPtrType {
		return data, nil
	}

	pc := &PathConfig{}
	if err := pc.DecodeMapstructure(data); err != nil {
		return nil, err
	}

	// Return pointer if target is pointer type
	if to == pathConfigPtrType {
		return pc, nil
	}
	return *pc, nil
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, ErrConfigPathEmpty
	}
	cfg, err := initConfig[Config](configPath)
	if err != nil {
		return nil, fmt.Errorf("initConfig[Config]: %w", err)
	}
	return cfg, nil
}
