package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level structure of an api.yaml file.
type Config struct {
	Version   string     `yaml:"version"`
	Info      Info       `yaml:"info"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Info struct {
	Title    string `yaml:"title"`
	BasePath string `yaml:"base_path"`
}

type Endpoint struct {
	Path     string    `yaml:"path"`
	Method   string    `yaml:"method"`
	Request  *Request  `yaml:"request,omitempty"`
	Behavior *Behavior `yaml:"behavior,omitempty"`
	Response *Response `yaml:"response,omitempty"`
	Variants []Variant `yaml:"variants,omitempty"`

	// source line number for error messages
	Line int `yaml:"-"`
}

type Request struct {
	RequiredFields []string               `yaml:"required_fields,omitempty"`
	BodySchema     map[string]FieldSchema `yaml:"body_schema,omitempty"`
}

type FieldSchema struct {
	Type      string `yaml:"type"`
	MinLength int    `yaml:"min_length,omitempty"`
	Format    string `yaml:"format,omitempty"`
}

type Behavior struct {
	ErrorRate   float64     `yaml:"error_rate,omitempty"`
	ErrorStatus int         `yaml:"error_status,omitempty"`
	ErrorBody   interface{} `yaml:"error_body,omitempty"`
	LatencyMs   int         `yaml:"latency_ms,omitempty"`
}

type Response struct {
	Status    int                 `yaml:"status"`
	Headers   map[string]string   `yaml:"headers,omitempty"`
	Body      interface{}         `yaml:"body,omitempty"`
	Count     int                 `yaml:"count,omitempty"`
	Paginated bool                `yaml:"paginated,omitempty"`
	Total     int                 `yaml:"total,omitempty"`
}

type Variant struct {
	Condition *Condition `yaml:"condition,omitempty"`
	Default   bool       `yaml:"default,omitempty"`
	Response  *Response  `yaml:"response"`
}

type Condition struct {
	Query   map[string]string `yaml:"query,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// Load reads, parses, and validates a YAML config file.
func Load(path string) (*Config, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("file %q not found", path)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, nil, fmt.Errorf("parse error: %w", err)
	}

	warnings, errs := validate(&cfg)
	if len(errs) > 0 {
		return nil, warnings, fmt.Errorf("%s", strings.Join(errs, "\n"))
	}

	normalize(&cfg)
	return &cfg, warnings, nil
}
