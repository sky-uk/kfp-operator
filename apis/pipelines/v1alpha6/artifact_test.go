//go:build unit

package v1alpha6

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"math/rand"
)

var _ = Context("OutputArtifact", func() {
	var _ = Describe("ArtifactPathFromString", func() {
		expectedArtifactPath := ArtifactPath{
			Locator: ArtifactLocator{
				Component: apis.RandomString(),
				Artifact:  apis.RandomString(),
				Index:     rand.Int(),
			},
			Filter: apis.RandomString(),
		}

		Specify("parses without filter and with index", func() {
			artifactPath, err := ArtifactPathFromString(fmt.Sprintf("%s:%s:%d", expectedArtifactPath.Locator.Component, expectedArtifactPath.Locator.Artifact, expectedArtifactPath.Locator.Index))
			Expect(err).NotTo(HaveOccurred())
			Expect(artifactPath.Locator).To(Equal(expectedArtifactPath.Locator))
			Expect(artifactPath.Filter).To(BeEmpty())
		})

		Specify("parses without filter and without index", func() {
			artifactPath, err := ArtifactPathFromString(fmt.Sprintf("%s:%s", expectedArtifactPath.Locator.Component, expectedArtifactPath.Locator.Artifact))
			Expect(err).NotTo(HaveOccurred())
			Expect(artifactPath.Locator.Component).To(Equal(expectedArtifactPath.Locator.Component))
			Expect(artifactPath.Locator.Artifact).To(Equal(expectedArtifactPath.Locator.Artifact))
			Expect(artifactPath.Locator.Index).To(Equal(0))
			Expect(artifactPath.Filter).To(BeEmpty())
		})

		Specify("parses with filter and with index", func() {
			artifactPath, err := ArtifactPathFromString(fmt.Sprintf("%s:%s:%d[%s]", expectedArtifactPath.Locator.Component, expectedArtifactPath.Locator.Artifact, expectedArtifactPath.Locator.Index, expectedArtifactPath.Filter))
			Expect(err).NotTo(HaveOccurred())
			Expect(artifactPath).To(Equal(expectedArtifactPath))
		})

		Specify("parses with filter and without index", func() {
			artifactPath, err := ArtifactPathFromString(fmt.Sprintf("%s:%s[%s]", expectedArtifactPath.Locator.Component, expectedArtifactPath.Locator.Artifact, expectedArtifactPath.Filter))
			Expect(err).NotTo(HaveOccurred())
			Expect(artifactPath.Locator.Component).To(Equal(expectedArtifactPath.Locator.Component))
			Expect(artifactPath.Locator.Artifact).To(Equal(expectedArtifactPath.Locator.Artifact))
			Expect(artifactPath.Locator.Index).To(Equal(0))
			Expect(artifactPath.Filter).To(Equal(expectedArtifactPath.Filter))
		})

		Specify("rejects missing component", func() {
			_, err := ArtifactPathFromString(fmt.Sprintf(":%s:%d[%s]", expectedArtifactPath.Locator.Artifact, expectedArtifactPath.Locator.Index, expectedArtifactPath.Filter))
			Expect(err).To(HaveOccurred())
		})

		Specify("rejects missing artifact", func() {
			_, err := ArtifactPathFromString(fmt.Sprintf("%s::%d[%s]", expectedArtifactPath.Locator.Component, expectedArtifactPath.Locator.Index, expectedArtifactPath.Filter))
			Expect(err).To(HaveOccurred())
		})

		Specify("rejects missing segments", func() {
			_, err := ArtifactPathFromString(fmt.Sprintf("%s[%s]", expectedArtifactPath.Locator.Component, expectedArtifactPath.Filter))
			Expect(err).To(HaveOccurred())
		})
	})
})
