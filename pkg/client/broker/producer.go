package broker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const flushTimeout = 5000 // milliseconds

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(address []string) (*Producer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(address, ","),
	}
	fmt.Println("Connecting to Kafka at:", strings.Join(address, ","))
	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, err
	}
	return &Producer{producer: p}, nil
}

func (p *Producer) ProducePartitionAny(message, topic string) error {
	kafkaMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(message),
		Key:   nil,
	}
	kafkaChannel := make(chan kafka.Event)
	if err := p.producer.Produce(kafkaMessage, kafkaChannel); err != nil {
		return err
	}
	e := <-kafkaChannel
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return ev
	default:
		return errors.New("unknown error")
	}

}

func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}
