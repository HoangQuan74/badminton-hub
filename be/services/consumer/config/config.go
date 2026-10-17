package consumerConfig

import (
	"fmt"
	"strings"
	"time"

	platformConfig "github.com/kainguyen/badminton-hub/be/platform/config"
)

type Config struct {
	KafkaBrokers      []string
	KafkaClientID     string
	KafkaGroupID      string
	KafkaDialTimeout  time.Duration
	KafkaReadTimeout  time.Duration
	KafkaWriteTimeout time.Duration
}

func Load() (Config, error) {
	brokers := splitBrokers(platformConfig.Env("CONSUMER_KAFKA_BROKERS", ""))
	if len(brokers) == 0 {
		return Config{}, fmt.Errorf("CONSUMER_KAFKA_BROKERS is required")
	}

	groupID := strings.TrimSpace(platformConfig.Env("CONSUMER_KAFKA_GROUP_ID", ""))
	if groupID == "" {
		return Config{}, fmt.Errorf("CONSUMER_KAFKA_GROUP_ID is required")
	}

	dialTimeout, err := platformConfig.Duration("CONSUMER_KAFKA_DIAL_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := platformConfig.Duration("CONSUMER_KAFKA_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := platformConfig.Duration("CONSUMER_KAFKA_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	return Config{
		KafkaBrokers:      brokers,
		KafkaClientID:     strings.TrimSpace(platformConfig.Env("CONSUMER_KAFKA_CLIENT_ID", "badminton-hub-consumer")),
		KafkaGroupID:      groupID,
		KafkaDialTimeout:  dialTimeout,
		KafkaReadTimeout:  readTimeout,
		KafkaWriteTimeout: writeTimeout,
	}, nil
}

func splitBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		if broker := strings.TrimSpace(part); broker != "" {
			brokers = append(brokers, broker)
		}
	}
	return brokers
}
