package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"dario.cat/mergo"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type EnvironmentProperties struct {
	HttpClient    *HttpClientProperties   `mapstructure:"httpClient"`
	Nodes         []*NodeProperties       `mapstructure:"nodes"`
	Broadcasting  *BroadcastingProperties `mapstructure:"broadcasting"`
	Filters       []*FilterProperties     `mapstructure:"filters"`
	Stores        []*StoreProperties      `mapstructure:"stores"`
	ProgramErrors ProgramErrors           `mapstructure:"programErrors"`
}

func LoadConfig[P any](ctx context.Context, applicationYml string, rootProperty string) (*P, error) {
	baseContent := replaceEnvVariables(applicationYml)
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(baseContent)); err != nil {
		return nil, fmt.Errorf("error reading embedded config: %w", err)
	}

	if rootProperty != "" {
		scoped := v.Sub(rootProperty)
		if scoped == nil {
			return nil, fmt.Errorf("root property %q not found in config", rootProperty)
		}
		v = scoped
	}

	var baseConfig P
	if err := v.Unmarshal(&baseConfig, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToWeakSliceHookFunc(","),
		))); err != nil {
		return nil, fmt.Errorf("error unmarshalling embedded config: %w", err)
	}
	slog.InfoContext(ctx, "Loaded embedded application.yml")

	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		files := splitAndTrim(configPath, ",")
		for _, file := range files {
			if !fileExists(file) {
				return nil, fmt.Errorf("config file not found: %s", file)
			}
			slog.DebugContext(ctx, "[CONFIG] merging config file", "filename", file)

			fileContent, err := os.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("error reading config file %s: %w", file, err)
			}

			overrideContent := replaceEnvVariables(string(fileContent))
			tmpV := viper.New()

			ext := filepath.Ext(file)
			switch ext {
			case ".yaml", ".yml":
				tmpV.SetConfigType("yaml")
			case ".json":
				tmpV.SetConfigType("json")
			case ".toml":
				tmpV.SetConfigType("toml")
			default:
				return nil, fmt.Errorf("unsupported config file extension: %s", ext)
			}

			if err := tmpV.ReadConfig(bytes.NewBufferString(overrideContent)); err != nil {
				return nil, fmt.Errorf("error reading config file %s: %w", file, err)
			}

			if rootProperty != "" {
				scoped := tmpV.Sub(rootProperty)
				if scoped == nil {
					return nil, fmt.Errorf("root property %q not found in override config %s", rootProperty, file)
				}
				tmpV = scoped
			}

			var overrideConfig P
			if err := tmpV.Unmarshal(&overrideConfig); err != nil {
				return nil, fmt.Errorf("error unmarshalling override config: %w", err)
			}

			// Deep merge override into base
			if err := mergo.Merge(&baseConfig, overrideConfig, mergo.WithOverride); err != nil {
				return nil, fmt.Errorf("error merging config: %w", err)
			}
			slog.InfoContext(ctx, "[CONFIG] Successfully merged config file", "filename", file)
		}
	}

	if jsonBytes, err := json.MarshalIndent(baseConfig, "", "  "); err == nil {
		slog.DebugContext(ctx, "[CONFIG] Final merged configuration", "config", string(jsonBytes))
	}

	return &baseConfig, nil
}

var envPattern = regexp.MustCompile(`\$\{(\w+)(?::([^}]*))?\}`)

func replaceEnvVariables(content string) string {
	return envPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := envPattern.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}
		envVar := submatches[1]
		defaultVal := submatches[2]
		if val, exists := os.LookupEnv(envVar); exists {
			return val
		}
		return defaultVal
	})
}

func splitAndTrim(s, sep string) []string {
	raw := bytes.Split([]byte(s), []byte(sep))
	var result []string
	for _, part := range raw {
		trimmed := string(bytes.TrimSpace(part))
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
