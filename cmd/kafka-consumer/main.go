package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"l0/internal/config"
	"l0/pkg/client/broker"
	"l0/pkg/logging"

	"github.com/IBM/sarama"
)

const logFile = "logs/kafka_consumer.log"

func main() {
	logging.InitLogger(logFile)
	logger, err := logging.GetLogger(logFile)
	if err != nil {
		panic(err)
	}
	cfg := config.GetConfig(logFile)

	topic := "order"
	msgCnt := 0

	worker, err := broker.ConnectConsumer(cfg.Brokers)
	if err != nil {
		panic(err)
	}

	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	consumerStr := broker.NewConsumer(consumer, logger)

	fmt.Println("Consumer started ")

	// 2. Handle OS signals - used to stop the process.
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// 3. Create a Goroutine to run the consumer / worker.
	doneCh := make(chan struct{})

	go consumerStr.RunConsumer(sigchan, &msgCnt, doneCh)

	<-doneCh
	fmt.Println("Processed", msgCnt, "messages")

	// 4. Close the consumer on exit.
	if err := worker.Close(); err != nil {
		panic(err)
	}

}

func ConnectConsumer(brokers []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	return sarama.NewConsumer(brokers, config)
}
