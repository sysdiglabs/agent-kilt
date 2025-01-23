package config

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

func decompressBytes(data []byte) []byte {
	reader, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		panic(fmt.Errorf("error constructing decompressor: %w", err))
	}

	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("error reading gzipped stream: %w", err))
	}
	return decompressedData
}
