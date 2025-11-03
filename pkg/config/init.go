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
	// Configure mapstructure hook for PathConfig decoding
	decoderConfig := &mapstructure.DecoderConfig{
		Result:           &cfg,
		DecodeHook:      pathConfigDecodeHook,
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, fmt.Errorf("mapstructure.NewDecoder: %w", err)
	}

	// Get raw config data
	raw := k.Raw()
	if err := decoder.Decode(raw); err != nil {
		return nil, fmt.Errorf("decoder.Decode: %w", err)
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
