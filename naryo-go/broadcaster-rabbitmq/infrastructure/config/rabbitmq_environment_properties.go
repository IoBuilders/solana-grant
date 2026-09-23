package config

type RabbitMQEnvironmentProperties struct {
	RabbitMQ *RabbitMQBroadcasterProperties `mapstructure:"rabbitmq"`
}
