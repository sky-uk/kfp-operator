package file

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
)

type FileHandler interface {
	Write(p []byte, location string, fileName string) error
	Delete(id string, location string) error
}

type GcsFileHandler struct {
	ctx       context.Context
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

	return GcsFileHandler{ctx: ctx, gcsClient: *client}, err
}

func (g *GcsFileHandler) Write(p []byte, location string, fileName string) error {
	writer := g.gcsClient.Bucket(location).Object(fileName).NewWriter(g.ctx)

	_, err := io.Writer(writer).Write(p)
	if err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}

	return nil
}

func (g *GcsFileHandler) Delete(id string, location string) error {
	query := &storage.Query{Prefix: fmt.Sprintf("%s/", id)}

	it := g.gcsClient.Bucket(location).Objects(g.ctx, query)
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return err
		}

		err = g.gcsClient.Bucket(location).Object(attrs.Name).Delete(g.ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
