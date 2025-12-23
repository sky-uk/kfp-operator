//go:build unit

package provider

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/option"
)

var _ = Describe("GcsFileHandler", Ordered, func() {
	var (
		ctx       = context.Background()
		server    *fakestorage.Server
		client    *storage.Client
		handler   GcsFileHandler
		bucket    string
		filePath  string
		testData  map[string]any
		testBytes []byte
	)

	BeforeAll(func() {
		var err error
		server, err = fakestorage.NewServerWithOptions(fakestorage.Options{
			InitialObjects: []fakestorage.Object{},
			NoListener:     true,
		})
		Expect(err).ShouldNot(HaveOccurred())

		ctx = context.Background()

		// Create a GCS client using the fake server's HTTP client
		// Use option.WithoutAuthentication() to avoid ADC conflicts
		client, err = storage.NewClient(ctx,
			option.WithHTTPClient(server.HTTPClient()),
			option.WithoutAuthentication())
		Expect(err).ShouldNot(HaveOccurred())

		// fake GCS server doesn't share data between different clients
		// (required for a round-trip test), so the same client must be used
		// throughout.
		handler = GcsFileHandler{*client}

		bucket = "test-bucket"
		filePath = "test-folder/test-file.json"
		testData = map[string]any{"key": "value"}
		testBytes, _ = json.Marshal(testData)

		err = client.Bucket(bucket).Create(ctx, "test-project", nil)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterAll(func() {
		server.Stop()
	})

	Context("Write, Read And Delete round trip", func() {
		When("Write", func() {
			It("should write data to the specified bucket and file path", func() {
				err := handler.Write(ctx, testBytes, bucket, filePath)
				err = handler.Write(ctx, testBytes, bucket, "test-folder/test-file2.json")
				Expect(err).ShouldNot(HaveOccurred())

				obj, err := client.Bucket(bucket).Object(filePath).Attrs(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(obj.Name).To(Equal(filePath))
			})
		})
		When("Read", func() {
			It("should extract the written data from the bucket", func() {
				readData, err := handler.Read(ctx, bucket, filePath)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(readData).To(Equal(testData))
			})
		})
		When("Delete", func() {
			It("should delete the file in the bucket", func() {
				err := handler.Delete(ctx, "test-folder", bucket)
				Expect(err).ShouldNot(HaveOccurred())

				readData, err := handler.Read(ctx, bucket, filePath)
				Expect(err).Should(HaveOccurred())
				Expect(readData).Should(BeNil())

				readData, err = handler.Read(ctx, bucket, "test-folder/test-file2.json")
				Expect(err).Should(HaveOccurred())
				Expect(readData).Should(BeNil())
			})
		})
	})
})
