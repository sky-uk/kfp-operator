//go:build unit

package mocks

import "github.com/stretchr/testify/mock"

type MockLabelSanitizer struct{ mock.Mock }

func (m *MockLabelSanitizer) Sanitize(labels map[string]string) map[string]string {
	args := m.Called(labels)
	var processedLabels map[string]string
	if arg0 := args.Get(0); arg0 != nil {
		processedLabels = arg0.(map[string]string)
	}
	return processedLabels
}
