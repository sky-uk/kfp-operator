//go:build unit

package mocks

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

type MockLabelGen struct{}

func (lg MockLabelGen) GenerateLabels(value any) (map[string]string, error) {
	switch v := value.(type) {
	case resource.RunDefinition:
		return map[string]string{
			"rd-key": "rd-value",
		}, nil
	case base.RunScheduleDefinition:
		return map[string]string{
			"rsd-key": "rsd-value",
		}, nil
	default:
		return nil, fmt.Errorf("Unexpected value of type %T", v)
	}
}
