package v1alpha6

import (
	"fmt"
)

var DefaultWorkflowNamespace string

func addDefaultWorkflowNamespaceToProvider(provider string) string {
	return fmt.Sprintf("%s/%s", DefaultWorkflowNamespace, provider)
}
