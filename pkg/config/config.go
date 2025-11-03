// Package config provides configuration structures for the OpenAPI filter tool.
// It defines the configuration format for filtering OpenAPI specs and
// tool-specific settings.
package config

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Config represents the root configuration structure for the OpenAPI filter tool.
// It combines tool-specific settings with filter configuration.
type Config struct {
	Tool         ToolConfig `koanf:"x-openapi-filter"`
	FilterConfig `koanf:",squash"`
}

// FilterConfig defines the configuration for filtering an OpenAPI spec.
// It specifies which parts of the spec should be included in the output.
type FilterConfig struct {
	Servers            bool                    `koanf:"servers"`             // Include servers section
	PreservePathServers bool                   `koanf:"preservePathServers"` // Preserve path-level servers (default: false)
	Paths              map[string]PathConfig   `koanf:"paths"`               // Map of paths to path configuration
	Components         *FilterComponentsConfig `koanf:"components"`          // Component filtering configuration
	Security           bool                    `koanf:"security"`            // Include security requirements
	Tags               bool                    `koanf:"tags"`                // Include tags
	ExternalDocs       bool                    `koanf:"externalDocs"`         // Include external documentation
}

// FilterComponentsConfig specifies which components should be included in the
// filtered OpenAPI spec. Each field is a list of component names to include.
type FilterComponentsConfig struct {
	Schemas         []string `koanf:"schemas"`         // List of schema names to include
	Parameters      []string `koanf:"parameters"`      // List of parameter names to include
	SecuritySchemes []string `koanf:"securitySchemes"` // List of security scheme names to include
	RequestBodies   []string `koanf:"requestBodies"`   // List of request body names to include
	Responses       []string `koanf:"responses"`       // List of response names to include
	Headers         []string `koanf:"headers"`         // List of header names to include
	Examples        []string `koanf:"examples"`        // List of example names to include
	Links           []string `koanf:"links"`           // List of link names to include
	Callbacks       []string `koanf:"callbacks"`       // List of callback names to include
}

// ToolConfig contains tool-specific configuration settings.
type ToolConfig struct {
	Logger *LoggerConfig `koanf:"logger"` // Logger configuration
	Loader *LoaderConfig `koanf:"loader"` // OpenAPI loader configuration
}

// LoggerConfig defines the logging configuration for the tool.
type LoggerConfig struct {
	Level string `koanf:"level"` // Log level (e.g., "debug", "info", "warn", "error")
}

// LoaderConfig defines configuration for the OpenAPI spec loader.
type LoaderConfig struct {
	IsExternalRefsAllowed bool `koanf:"external_refs_allowed"` // Whether to allow external references
}

// PathConfig defines configuration for a single API path.
// It supports both simple format (array of methods) and advanced format (object with methods and preserveServers).
type PathConfig struct {
	Methods         []string `koanf:"methods"`         // List of HTTP methods to include
	PreserveServers bool     `koanf:"preserveServers"` // Whether to preserve path-level servers
}

// UnmarshalJSON implements custom JSON unmarshaling to support both simple array format
// (backward compatible) and advanced object format.
func (pc *PathConfig) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first (simple format)
	var methods []string
	if err := json.Unmarshal(data, &methods); err == nil {
		pc.Methods = methods
		pc.PreserveServers = false
		return nil
	}

	// Try to unmarshal as object (advanced format)
	var obj struct {
		Methods         []string `json:"methods"`
		PreserveServers bool     `json:"preserveServers"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		pc.Methods = obj.Methods
		pc.PreserveServers = obj.PreserveServers
		return nil
	}

	return fmt.Errorf("invalid path config format: expected array of strings or object with methods field")
}

// UnmarshalYAML implements custom YAML unmarshaling to support both simple array format
// (backward compatible) and advanced object format.
func (pc *PathConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as array first (simple format)
	var methods []string
	if err := unmarshal(&methods); err == nil {
		pc.Methods = methods
		pc.PreserveServers = false
		return nil
	}

	// Try to unmarshal as object (advanced format)
	var obj struct {
		Methods         []string `yaml:"methods"`
		PreserveServers bool     `yaml:"preserveServers"`
	}
	if err := unmarshal(&obj); err == nil {
		pc.Methods = obj.Methods
		pc.PreserveServers = obj.PreserveServers
		return nil
	}

	return fmt.Errorf("invalid path config format: expected array of strings or object with methods field")
}

// DecodeMapstructure implements custom decoding for mapstructure (used by koanf).
// This allows koanf to properly decode PathConfig from raw interface{} values.
func (pc *PathConfig) DecodeMapstructure(from interface{}) error {
	if from == nil {
		return fmt.Errorf("path config cannot be nil")
	}

	val := reflect.ValueOf(from)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		// Simple format: array of strings
		methods := make([]string, val.Len())
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if elem.Kind() == reflect.String {
				methods[i] = elem.String()
			} else if elem.Kind() == reflect.Interface {
				if str, ok := elem.Interface().(string); ok {
					methods[i] = str
				} else {
					return fmt.Errorf("path config array element must be a string, got %T", elem.Interface())
				}
			} else {
				return fmt.Errorf("path config array element must be a string, got %v", elem.Kind())
			}
		}
		pc.Methods = methods
		pc.PreserveServers = false
		return nil

	case reflect.Map:
		// Advanced format: map with methods and optional preserveServers
		pc.Methods = []string{}
		pc.PreserveServers = false

		iter := val.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			if key.Kind() == reflect.String {
				switch key.String() {
				case "methods":
					// Handle interface{} wrapping
					actualValue := value
					if value.Kind() == reflect.Interface {
						actualValue = reflect.ValueOf(value.Interface())
					}
					if actualValue.Kind() == reflect.Slice || actualValue.Kind() == reflect.Array {
						methods := make([]string, actualValue.Len())
						for i := 0; i < actualValue.Len(); i++ {
							elem := actualValue.Index(i)
							if elem.Kind() == reflect.String {
								methods[i] = elem.String()
							} else if elem.Kind() == reflect.Interface {
								if str, ok := elem.Interface().(string); ok {
									methods[i] = str
								} else {
									return fmt.Errorf("methods array element must be a string, got %T", elem.Interface())
								}
							} else {
								return fmt.Errorf("methods array element must be a string, got %v", elem.Kind())
							}
						}
						pc.Methods = methods
					} else {
						return fmt.Errorf("methods field must be an array, got %v", actualValue.Kind())
					}
				case "preserveServers":
					if value.Kind() == reflect.Bool {
						pc.PreserveServers = value.Bool()
					} else if value.Kind() == reflect.Interface {
						if b, ok := value.Interface().(bool); ok {
							pc.PreserveServers = b
						} else {
							return fmt.Errorf("preserveServers field must be a boolean, got %T", value.Interface())
						}
					} else {
						return fmt.Errorf("preserveServers field must be a boolean, got %v", value.Kind())
					}
				}
			}
		}

		if len(pc.Methods) == 0 {
			return fmt.Errorf("path config object must have a methods field")
		}

		return nil
	}

	return fmt.Errorf("invalid path config format: expected array or object, got %v", val.Kind())
}
