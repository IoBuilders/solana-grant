package config

// EnvironmentProperties is the server module's own slice of the shared naryo.* YAML tree: just
// the HTTP server settings, loaded the same way core and persistence-gorm load theirs.
type EnvironmentProperties struct {
	Server *ServerProperties `mapstructure:"server"`
}

type ServerProperties struct {
	Port int `mapstructure:"port"`
}
