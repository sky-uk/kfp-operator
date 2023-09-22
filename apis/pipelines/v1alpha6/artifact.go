package v1alpha6

import (
	"fmt"
	"regexp"
	"strconv"
)

type ArtifactLocator struct {
	Component string `json:"-" yaml:"-"`
	Artifact  string `json:"-" yaml:"-"`
	Index     int    `json:"-" yaml:"-"`
}

func (ap *ArtifactLocator) String() string {
	return fmt.Sprintf("%s:%s:%d", ap.Component, ap.Artifact, ap.Index)
}

const ArtifactPathPattern = `^([^\[\]:]+):([^\[\]:]+)(?::(\d*))?(?:\[([^\[\]:]+)\])?$`

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Pattern=`^([^\[\]:]+):([^\[\]:]+)(?::(\d*))?(?:\[([^\[\]:]+)\])?$`
type ArtifactPath struct {
	Locator ArtifactLocator `json:"-" yaml:"-"`
	Filter  string          `json:"-" yaml:"-"`
}

func (ap ArtifactPath) String() string {
	if ap.Filter == "" {
		return ap.Locator.String()
	}

	return fmt.Sprintf("%s[%s]", ap.Locator.String(), ap.Filter)
}

func ArtifactPathFromString(path string) (artifactPath ArtifactPath, err error) {
	pathPattern := regexp.MustCompile(ArtifactPathPattern)
	matches := pathPattern.FindStringSubmatch(path)

	if len(matches) < 3 {
		err = fmt.Errorf("ArtifactPath must match pattern %s", ArtifactPathPattern)
		return
	}

	artifactPath.Locator = ArtifactLocator{
		Component: matches[1],
		Artifact:  matches[2],
	}

	if len(matches) > 3 && matches[3] != "" {
		var index int
		index, err = strconv.Atoi(matches[3])
		if err != nil {
			return
		}

		artifactPath.Locator.Index = index
	}

	if len(matches) > 4 {
		artifactPath.Filter = matches[4]
	}

	return
}

func (ap ArtifactPath) MarshalText() ([]byte, error) {
	return []byte(ap.String()), nil
}

func (ap *ArtifactPath) UnmarshalText(bytes []byte) error {
	deserialised, err := ArtifactPathFromString(string(bytes))
	*ap = deserialised

	return err
}

type OutputArtifact struct {
	Name string       `json:"name"`
	Path ArtifactPath `json:"path"`
}

func (oa OutputArtifact) GetKey() string {
	return oa.Name
}

func (oa OutputArtifact) GetValue() string {
	return oa.Path.String()
}
