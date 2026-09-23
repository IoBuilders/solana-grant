package config

type KafkaEnvironmentProperties struct {
	Kafka *KafkaBroadcasterProperties `mapstructure:"kafka"`
}
