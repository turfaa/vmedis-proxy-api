package zstd2

import (
	"bytes"

	"github.com/klauspost/compress/zstd"
)

func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	w, err := zstd.NewWriter(&buf, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	r, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
