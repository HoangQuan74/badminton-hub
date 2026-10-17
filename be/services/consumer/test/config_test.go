package consumertest

import (
	"testing"

	consumerConfig "github.com/kainguyen/badminton-hub/be/services/consumer/config"
)

func TestLoadKafkaConfig(t *testing.T) {
	t.Setenv("CONSUMER_KAFKA_BROKERS", " kafka-1:9092, kafka-2:9092 ")
	t.Setenv("CONSUMER_KAFKA_CLIENT_ID", "consumer-test")
	t.Setenv("CONSUMER_KAFKA_GROUP_ID", "booking-consumer")

	cfg, err := consumerConfig.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.KafkaBrokers) != 2 || cfg.KafkaBrokers[0] != "kafka-1:9092" {
		t.Fatalf("KafkaBrokers = %#v", cfg.KafkaBrokers)
	}
	if cfg.KafkaGroupID != "booking-consumer" {
		t.Fatalf("KafkaGroupID = %q", cfg.KafkaGroupID)
	}
}

func TestLoadKafkaConfigRequiresBrokers(t *testing.T) {
	t.Setenv("CONSUMER_KAFKA_BROKERS", "")
	t.Setenv("CONSUMER_KAFKA_GROUP_ID", "booking-consumer")

	if _, err := consumerConfig.Load(); err == nil {
		t.Fatal("Load() error = nil, want missing brokers error")
	}
}
