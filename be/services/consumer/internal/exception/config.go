package exception

import "errors"

var (
	ErrKafkaBrokersRequired = errors.New("CONSUMER_KAFKA_BROKERS is required")
	ErrKafkaGroupIDRequired = errors.New("CONSUMER_KAFKA_GROUP_ID is required")
)
