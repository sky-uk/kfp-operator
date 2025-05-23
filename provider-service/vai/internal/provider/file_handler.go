package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FileHandler interface {
	Write(ctx context.Context, content []byte, bucket string, filePath string) error
	Delete(ctx context.Context, id string, bucket string) error
	Read(ctx context.Context, bucket string, filePath string) (map[string]any, error)
}

type GcsFileHandler struct {
	gcsClient storage.Client
}

func NewGcsFileHandler(
	ctx context.Context,
	gcsEndpoint string,
) (GcsFileHandler, error) {
	var client *storage.Client
	var err error
	if gcsEndpoint != "" {
		client, err = storage.NewClient(
			ctx,
			option.WithoutAuthentication(),
			option.WithEndpoint(gcsEndpoint),
		)
	} else {
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return GcsFileHandler{}, err
	}

	return GcsFileHandler{gcsClient: *client}, nil
}

// Write writes bytes into the location inferred by GCS bucket name and
// file path (relative to GCS bucket location).
func (g *GcsFileHandler) Write(
	ctx context.Context,
	content []byte,
	bucket string,
	filePath string,
) error {
	writer := g.gcsClient.Bucket(bucket).Object(filePath).NewWriter(ctx)

	_, err := io.Writer(writer).Write(content)
	if err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	return nil
}

// Delete deletes all files inferred by the GCS bucket name and id.
func (g *GcsFileHandler) Delete(ctx context.Context, id string, bucket string) error {
	query := &storage.Query{Prefix: fmt.Sprintf("%s/", id)}

	it := g.gcsClient.Bucket(bucket).Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return err
		}

		err = g.gcsClient.Bucket(bucket).Object(attrs.Name).Delete(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Read reads and returns the unmarshalled from the location inferred by the
// GCS bucket name and file path.
func (g *GcsFileHandler) Read(
	ctx context.Context,
	bucket string,
	filePath string,
) (map[string]any, error) {
	reader, err := g.gcsClient.Bucket(bucket).Object(filePath).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	raw := map[string]any{}
	err = json.Unmarshal(buf.Bytes(), &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
