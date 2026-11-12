package vmedisv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// number is a JSON number that vmedis sends either as a number (6500) or as
// a string ("6500.00"). null and "" decode to 0.
type number float64

func (n *number) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*n = 0
		return nil
	}

	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s == "" {
			*n = 0
			return nil
		}

		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("parse number %q: %w", s, err)
		}
		*n = number(f)
		return nil
	}

	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*n = number(f)
	return nil
}
