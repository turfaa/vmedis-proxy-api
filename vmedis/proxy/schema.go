package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// Stock represents one instance of stock.
type Stock struct {
	Unit     string  `json:"unit"`
	Quantity float64 `json:"quantity"`
}

// MarshalText implements encoding.TextMarshaler.
func (s Stock) MarshalText() ([]byte, error) {
	q, err := json.Marshal(s.Quantity)
	if err != nil {
		return nil, fmt.Errorf("marshal quantity: %w", err)
	}

	var b bytes.Buffer
	b.Write(q)
	if s.Unit != "" {
		b.WriteByte(' ')
		b.WriteString(s.Unit)
	}

	return bytes.TrimSpace(b.Bytes()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Stock) UnmarshalText(text []byte) error {
	split := bytes.SplitN(bytes.TrimSpace(text), []byte(" "), 2)

	var q float64
	if err := json.Unmarshal(split[0], &q); err != nil {
		return fmt.Errorf("unmarshal quantity: %w", err)
	}

	s.Quantity = q
	if len(split) > 1 {
		s.Unit = string(split[1])
	}

	return nil
}

// FromClientStock converts Stock from client schema to proxy schema.
func FromClientStock(cs client.Stock) Stock {
	return Stock{
		Unit:     cs.Unit,
		Quantity: cs.Quantity,
	}
}

// FromModelsStock converts Stock from models.Stock to proxy schema.
func FromModelsStock(stock models.Stock) Stock {
	return Stock{
		Unit:     stock.Unit,
		Quantity: stock.Quantity,
	}
}
