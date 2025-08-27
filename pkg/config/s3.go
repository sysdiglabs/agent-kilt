package config

import (
	"fmt"
	"io"
	"strings"

	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func FromS3(path string, decompress bool) string {

	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		panic(fmt.Errorf("could not load default aws config: %w", err))
	}

	splitPath := strings.SplitN(path, "/", 2)

	if len(splitPath) != 2 {
		panic(fmt.Errorf("invalid path specified: expected bucket/objectkey, got '%s'", path))
	}

	svc := s3.NewFromConfig(cfg)

	obj, err := svc.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &splitPath[0],
		Key:    &splitPath[1],
	})

	if err != nil {
		panic(fmt.Errorf("could not retrieve %s: %w", path, err))
	}

	config, err := io.ReadAll(obj.Body)
	defer obj.Body.Close()

	if err != nil {
		panic(fmt.Errorf("could not read %s: %w", path, err))
	}

	if decompress {
		config = decompressBytes(config)
	}

	return string(config)
}
