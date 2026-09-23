package config

type EnvironmentProperties struct {
	Database *DatabaseProperties `mapstructure:"database"`
}
