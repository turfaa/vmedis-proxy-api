package drug

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/jsonpb"
	"github.com/segmentio/kafka-go"

	"github.com/turfaa/vmedis-proxy-api/kafkapb"
)

const (
	VmedisIDUpdatedTopic   = "drug_vmedis_id.updated"
	VmedisCodeUpdatedTopic = "drug_vmedis_code.updated"
)

var (
	jsonpbMarshaler = jsonpb.Marshaler{Indent: "  "}
)

type Producer struct {
	writer *kafka.Writer
}

func (p *Producer) ProduceUpdatedDrugsByVmedisID(ctx context.Context, messages []*kafkapb.UpdatedDrugByVmedisID) error {
	kafkaMessages := make([]kafka.Message, 0, len(messages))
	for _, message := range messages {
		messageStr, err := jsonpbMarshaler.MarshalToString(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		kafkaMessages = append(kafkaMessages, kafka.Message{
			Topic: VmedisIDUpdatedTopic,
			Key:   []byte(strconv.FormatInt(message.VmedisId, 10)),
			Value: []byte(messageStr),
		})
	}

	if err := p.writer.WriteMessages(ctx, kafkaMessages...); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

func (p *Producer) ProduceUpdatedDrugByVmedisCode(ctx context.Context, messages []*kafkapb.UpdatedDrugByVmedisCode) error {
	kafkaMessages := make([]kafka.Message, 0, len(messages))
	for _, message := range messages {
		messageStr, err := jsonpbMarshaler.MarshalToString(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		kafkaMessages = append(kafkaMessages, kafka.Message{
			Topic: VmedisCodeUpdatedTopic,
			Key:   []byte(message.VmedisCode),
			Value: []byte(messageStr),
		})
	}

	if err := p.writer.WriteMessages(ctx, kafkaMessages...); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

func NewProducer(writer *kafka.Writer) *Producer {
	return &Producer{
		writer: writer,
	}
}
