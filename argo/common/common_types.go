package common

import (
	"encoding/json"
	"fmt"
	"strings"
)

type NamespacedName struct {
	Name      string `json:"-"`
	Namespace string `json:"-"`
}

func (nsn NamespacedName) string() (string, error) {
	if nsn.Namespace == "" {
		return nsn.Name, nil
	}

	if nsn.Name == ""  {
		return "", fmt.Errorf("namespace provided without a name")
	}

	return strings.Join([]string{nsn.Namespace, nsn.Name}, "/"), nil
}

func namespacedNameFromString(namespacedName string) (NamespacedName, error) {
	splits := strings.Split(namespacedName, "/")

	if len(splits) < 2 {
		return NamespacedName{
			Name: namespacedName,
		}, nil
	}

	if len(splits) > 2 {
		return NamespacedName{}, fmt.Errorf("NamespacedName must be separated by at most one `/`")
	}

	if splits[0] == "" || splits[1] == "" {
		return NamespacedName{}, fmt.Errorf("name and namespace must not be empty when separated by `/`")
	}

	return NamespacedName{
		Namespace: splits[0],
		Name:      splits[1],
	}, nil
}

func (nsn NamespacedName) MarshalJSON() ([]byte, error) {
	serialised, err := nsn.string()
	if err != nil {
		return nil, err
	}

	return json.Marshal(serialised)
}

func (nsn NamespacedName) UnmarshalJSON(bytes []byte) error {
	var pidStr string
	err := json.Unmarshal(bytes, &pidStr)
	if err != nil {
		return err
	}

	nsn, err = namespacedNameFromString(pidStr)

	return err
}
