package broker

import (
	"l0/pkg/logging"
	"os"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.PartitionConsumer
	logger   *logging.Logger
}

func NewConsumer(c sarama.PartitionConsumer, l *logging.Logger) *Consumer {
	return &Consumer{consumer: c, logger: l}
}

func (c Consumer) RunConsumer(sigchan chan os.Signal, msgCnt *int, doneCh chan struct{}) {
	for {
		select {
		case err := <-c.consumer.Errors():
			c.logger.Error(err)
		case msg := <-c.consumer.Messages():
			*msgCnt++
			c.logger.Infof("Received order Count %d: | Topic(%s) | Message(%s) \n", *msgCnt, string(msg.Topic), string(msg.Value))
		case <-sigchan:
			c.logger.Info("Interrupt is detected")
			doneCh <- struct{}{}
		}
	}
}

func ConnectConsumer(brokers []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	return sarama.NewConsumer(brokers, config)
}

// const logFile = "logs/kafka_consumer.log"

// broker.RunConsumer(cfg.Brokers, "order", *logger)

// func main() {
// 	logging.InitLogger(logFile)
// 	logger, err := logging.GetLogger(logFile)
// 	if err != nil {
// 		panic(err)
// 	}
// 	cfg := config.GetConfig(logFile)

// 	topic := "order"
// 	msgCnt := 0

// 	// 1. Create a new consumer and start it.
// 	worker, err := ConnectConsumer(cfg.Brokers)
// 	if err != nil {
// 		panic(err)
// 	}

// 	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("Consumer started ")

// 	// 2. Handle OS signals - used to stop the process.
// 	sigchan := make(chan os.Signal, 1)
// 	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

// 	// 3. Create a Goroutine to run the consumer / worker.
// 	doneCh := make(chan struct{})
// 	go func() {
// 		for {
// 			select {
// 			case err := <-consumer.Errors():
// 				fmt.Println(err)
// 			case msg := <-consumer.Messages():
// 				msgCnt++
// 				fmt.Printf("Received order Count %d: | Topic(%s) | Message(%s) \n", msgCnt, string(msg.Topic), string(msg.Value))
// 				order := string(msg.Value)
// 				fmt.Printf("Brewing coffee for order: %s\n", order)
// 			case <-sigchan:
// 				fmt.Println("Interrupt is detected")
// 				doneCh <- struct{}{}
// 			}
// 		}
// 	}()

// 	<-doneCh
// 	fmt.Println("Processed", msgCnt, "messages")

// 	// 4. Close the consumer on exit.
// 	if err := worker.Close(); err != nil {
// 		panic(err)
// 	}

// }
