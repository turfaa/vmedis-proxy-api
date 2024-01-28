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
	VmedisIDUpdated = "drug_vmedis_id.updated"
)

var (
	jsonpbMarshaler = jsonpb.Marshaler{Indent: "  "}
)

type Producer struct {
	writer *kafka.Writer
}

func (p *Producer) ProduceUpdatedDrug(ctx context.Context, requestKey string, vmedisID int64) error {
	message := kafkapb.UpdatedDrugByVmedisID{
		RequestKey: requestKey,
		VmedisId:   vmedisID,
	}

	messageStr, err := jsonpbMarshaler.MarshalToString(&message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: VmedisIDUpdated,
		Key:   []byte(strconv.FormatInt(vmedisID, 10)),
		Value: []byte(messageStr),
	}); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

func NewProducer(writer *kafka.Writer) *Producer {
	return &Producer{
		writer: writer,
	}
}
