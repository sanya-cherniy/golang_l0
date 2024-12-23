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
