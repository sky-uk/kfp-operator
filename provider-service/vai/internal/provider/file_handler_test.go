//go:build unit

package provider

import (
	"context"
	"encoding/json"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

var _ = Describe("GcsFileHandler", Ordered, func() {
	var (
		ctx              = context.Background()
		server           *fakestorage.Server
		handler          GcsFileHandler
		bucket           string
		filePath         string
		compiledPipeline resource.CompiledPipeline
	)

	BeforeAll(func() {
		var err error
		server = fakestorage.NewServer(nil)
		Expect(err).ShouldNot(HaveOccurred())

		ctx = context.Background()

		// fake GCS server doesn't share data between different clients
		// (required for a round-trip test), so the same client must be used
		// throughout.
		handler = GcsFileHandler{*server.Client()}
		Expect(err).ShouldNot(HaveOccurred())

		bucket = "test-bucket"
		filePath = "test-folder/test-file.json"
		testBytes, _ := json.Marshal(map[string]any{"key": "value"})
		compiledPipeline = resource.CompiledPipeline{
			DisplayName:  "display-name",
			Labels:       map[string]string{"label-key": "label-value"},
			PipelineSpec: testBytes,
		}

		err = server.Client().Bucket(bucket).Create(ctx, "test-project", nil)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterAll(func() {
		server.Stop()
	})

	Context("Write, Read And Delete round trip", func() {
		When("Write", func() {
			It("should write data to the specified bucket and file path", func() {
				err := handler.Write(ctx, compiledPipeline, bucket, filePath)
				Expect(err).ShouldNot(HaveOccurred())

				err = handler.Write(ctx, compiledPipeline, bucket, "test-folder/test-file2.json")
				Expect(err).ShouldNot(HaveOccurred())

				obj, err := server.Client().Bucket(bucket).Object(filePath).Attrs(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(obj.Name).To(Equal(filePath))
			})
		})
		When("Read", func() {
			It("should extract the written data from the bucket", func() {
				readData, err := handler.Read(ctx, bucket, filePath)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(readData).To(Equal(compiledPipeline))
			})
		})
		When("Delete", func() {
			It("should delete the file in the bucket", func() {
				err := handler.Delete(ctx, "test-folder", bucket)
				Expect(err).ShouldNot(HaveOccurred())

				readData, err := handler.Read(ctx, bucket, filePath)
				Expect(err).Should(HaveOccurred())
				Expect(readData).Should(Equal(resource.CompiledPipeline{}))

				readData, err = handler.Read(ctx, bucket, "test-folder/test-file2.json")
				Expect(err).Should(HaveOccurred())
				Expect(readData).Should(Equal(resource.CompiledPipeline{}))
			})
		})
	})
})
