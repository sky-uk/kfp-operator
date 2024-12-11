//go:build unit

package sources

import (
	"cloud.google.com/go/pubsub"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Pub sub source", func() {
	Describe("extractPipelineJobId", func() {
		source := PubSubSource{}

		It("extracts pipeline_job_id from a message", func() {
			expectedId := "some_id"

			entry := []byte(`
				{
				 "resource" : {
						"labels": {
							"pipeline_job_id" : "` + expectedId + `"
						}
					}
  				}
			`)

			pubsubMessage := &pubsub.Message{
				Data: entry,
			}

			result, err := source.extractPipelineJobId(pubsubMessage)
			Expect(err).ToNot(HaveOccurred())

			Expect(result).To(Equal(expectedId))
		})

		When("unmarshalling fails", func() {
			It("returns an error", func() {
				entry := []byte(`{/}`)
				pubsubMessage := &pubsub.Message{
					Data: entry,
				}

				result, err := source.extractPipelineJobId(pubsubMessage)
				Expect(result).To(Equal(""))
				Expect(err).To(HaveOccurred())
			})
		})

		When("message does not contain the pipeline_job_id label", func() {
			It("returns an error", func() {
				entry := []byte(`
				{
					"resource" : {
						"labels": {
							"some_other_label" : "some_other_value"
						}
					}
				}`)

				pubsubMessage := &pubsub.Message{
					Data: entry,
				}

				logEntry := LogEntry{Resource: Resource{
					Labels: map[string]string{
						"some_other_label": "some_other_value",
					},
				}}

				result, err := source.extractPipelineJobId(pubsubMessage)

				expectedErr := fmt.Errorf("logEntry did not contain pipeline_job_id in %+v", logEntry)

				Expect(result).To(Equal(""))
				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})
