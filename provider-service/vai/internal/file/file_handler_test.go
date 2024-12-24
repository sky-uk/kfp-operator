//go:build unit

package file

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGcsFileHandlerUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VAI GCS File Handler Unit Suite")
}

var _ = Describe("GcsFileHandler", func() {
	var (
		server  *fakestorage.Server
		handler GcsFileHandler
		ctx     context.Context

		bucketName string
		filePath   string
		testData   map[string]any
		testBytes  []byte
	)

	BeforeEach(func() {
		// Start the fake GCS server
		// server = fakestorage.NewServer(nil)
		server, err := fakestorage.NewServerWithOptions(
			fakestorage.Options{
				Scheme: "http",
			},
		)
		Expect(err).ShouldNot(HaveOccurred())

		ctx = context.Background()

		handler, err = NewGcsFileHandler(ctx, server.URL())
		Expect(err).ShouldNot(HaveOccurred())
		// Test data setup
		bucketName = "test-bucket"
		filePath = "test-folder/test-file.json"
		testData = map[string]any{"key": "value"}
		testBytes, _ = json.Marshal(testData)

		// Create a bucket
		err = server.Client().Bucket(bucketName).Create(ctx, "test-project", nil)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Stop()
	})

	Describe("Write", func() {
		It("should write data to the specified bucket and file path", func() {
			err := handler.Write(testBytes, bucketName, filePath)
			Expect(err).ShouldNot(HaveOccurred())

			// Verify the file exists in the fake GCS server
			obj, err := server.Client().Bucket(bucketName).Object(filePath).Attrs(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(obj.Name).To(Equal(filePath))
		})
	})

	// Describe("Read", func() {
	// 	BeforeEach(func() {
	// 		// Preload the object into the fake GCS server
	// 		writer := server.Client().Bucket(bucketName).Object(filePath).NewWriter(ctx)
	// 		_, err := io.WriteString(writer, string(testBytes))
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		err = writer.Close()
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 	})
	//
	// 	It("should read data from the specified bucket and file path", func() {
	// 		readData, err := handler.Read(bucketName, filePath)
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		Expect(readData).To(Equal(testData))
	// 	})
	// })
	//
	// Describe("Delete", func() {
	// 	BeforeEach(func() {
	// 		// Preload the object into the fake GCS server
	// 		writer := server.Client().Bucket(bucketName).Object(filePath).NewWriter(ctx)
	// 		_, err := io.WriteString(writer, string(testBytes))
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		err = writer.Close()
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 	})
	//
	// 	It("should delete objects with the specified prefix", func() {
	// 		err := handler.Delete("test-folder", bucketName)
	// 		Expect(err).ShouldNot(HaveOccurred())
	//
	// 		// Verify the file no longer exists
	// 		_, err = server.Client().Bucket(bucketName).Object(filePath).Attrs(ctx)
	// 		Expect(err).Should(HaveOccurred())
	// 	})
	// })
})
