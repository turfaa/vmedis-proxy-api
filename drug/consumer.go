package drug

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"
)

const (
	ConsumerGroupID = "drug-consumer"
)

type Consumer struct {
	brokers     []string
	handler     *ConsumerHandler
	readers     []*kafka.Reader
	lock        sync.Mutex
	concurrency int
}

func (c *Consumer) StartConsuming() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := c.StartConsumingDumpDrugDetailsByVmedisCode(); err != nil {
			log.Printf("StartConsumingDumpDrugDetailsByVmedisCode returns error: %s", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := c.StartConsumingDumpDrugDetailsByVmedisID(); err != nil {
			log.Printf("StartConsumingDumpDrugDetailsByVmedisID returns error: %s", err)
		}
	}()

	wg.Wait()
}

func (c *Consumer) StartConsumingDumpDrugDetailsByVmedisCode() error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.brokers,
		Topic:   VmedisCodeUpdatedTopic,
		GroupID: ConsumerGroupID,
	})

	c.lock.Lock()
	c.readers = append(c.readers, reader)
	c.lock.Unlock()

	messageChan := make(chan kafka.Message, 1000)

	var wg sync.WaitGroup
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.processDumpDrugDetailsByVmedisCode(messageChan)
		}()
	}

	defer func() {
		close(messageChan)
		wg.Wait()
	}()

	for {
		m, err := reader.ReadMessage(context.Background())
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		messageChan <- m
	}
}

func (c *Consumer) processDumpDrugDetailsByVmedisCode(messages <-chan kafka.Message) {
	for m := range messages {
		if err := c.handler.DumpDrugDetailsByVmedisCode(context.Background(), m); err != nil {
			log.Printf("failed to dump drug details by vmedis code: %s", err)
		}
	}
}

func (c *Consumer) StartConsumingDumpDrugDetailsByVmedisID() error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.brokers,
		Topic:   VmedisIDUpdatedTopic,
		GroupID: ConsumerGroupID,
	})

	c.lock.Lock()
	c.readers = append(c.readers, reader)
	c.lock.Unlock()

	messageChan := make(chan kafka.Message, 1000)

	var wg sync.WaitGroup
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.processDumpDrugDetailsByVmedisID(messageChan)
		}()
	}

	defer func() {
		close(messageChan)
		wg.Wait()
	}()

	for {
		m, err := reader.ReadMessage(context.Background())
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		messageChan <- m
	}
}

func (c *Consumer) processDumpDrugDetailsByVmedisID(messages <-chan kafka.Message) {
	for m := range messages {
		if err := c.handler.DumpDrugDetailsByVmedisID(context.Background(), m); err != nil {
			log.Printf("failed to dump drug details by vmedis id: %s", err)
		}
	}
}

func (c *Consumer) Close() {
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			log.Printf("failed to close reader: %s", err)
		}
	}
}

func NewConsumer(config ConsumerConfig) *Consumer {
	return &Consumer{
		brokers:     config.Brokers,
		handler:     NewConsumerHandler(config.DB, config.RedisClient, config.VmedisClient, config.KafkaWriter),
		concurrency: config.Concurrency,
	}
}
