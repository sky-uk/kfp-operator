package common

import (
	"fmt"
	"strings"
)

type NamespacedName struct {
	Name      string `json:"-" yaml:"-"`
	Namespace string `json:"-" yaml:"-"`
}

func (nsn NamespacedName) Empty() bool {
	return nsn.Name == "" && nsn.Namespace == ""
}

func (nsn NamespacedName) String() (string, error) {
	if nsn.Namespace == "" {
		return nsn.Name, nil
	}

	if nsn.Name == ""  {
		return "", fmt.Errorf("namespace provided without a name")
	}

	return strings.Join([]string{nsn.Namespace, nsn.Name}, "/"), nil
}

func NamespacedNameFromString(namespacedName string) (NamespacedName, error) {
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

func (nsn NamespacedName) MarshalText() ([]byte, error) {
	serialised, err := nsn.String()
	if err != nil {
		return nil, err
	}

	return []byte(serialised), nil
}

func (nsn *NamespacedName) UnmarshalText(bytes []byte) error {
	deserialised, err := NamespacedNameFromString(string(bytes))
	*nsn = deserialised

	return err
}
