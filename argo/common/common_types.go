package common

import (
	"fmt"
	"strings"
)

// +kubebuilder:validation:Type=string
type NamespacedName struct {
	Name      string `json:"-" yaml:"-"`
	Namespace string `json:"-" yaml:"-"`
}

func (nsn NamespacedName) Empty() bool {
	return nsn.Name == "" && nsn.Namespace == ""
}

func (nsn NamespacedName) SeparatedString(separator string) (string, error) {
	if nsn.Namespace == "" {
		return nsn.Name, nil
	}

	if nsn.Name == "" {
		return "", fmt.Errorf("namespace [%s] provided without a name", nsn.Namespace)
	}

	return strings.Join([]string{nsn.Namespace, nsn.Name}, separator), nil
}

func (nsn NamespacedName) String() (string, error) {
	return nsn.SeparatedString("/")
}

func NamespacedNameFromString(namespacedName string) (NamespacedName, error) {
	splits := strings.Split(namespacedName, "/")

	if len(splits) < 2 {
		return NamespacedName{
			Name: namespacedName,
		}, nil
	}

	if len(splits) > 2 {
		return NamespacedName{}, fmt.Errorf("NamespacedName [%s] must be separated by at most one `/`", namespacedName)
	}

	if splits[0] == "" || splits[1] == "" {
		return NamespacedName{}, fmt.Errorf("name and namespace [%s] must not be empty when separated by `/`", namespacedName)
	}

	return NamespacedName{
		Namespace: splits[0],
		Name:      splits[1],
	}, nil
}

// Optionally render structs until https://github.com/golang/go/issues/11939 is addressed
func (nsn NamespacedName) NonEmptyPtr() *NamespacedName {
	if nsn.Empty() {
		return nil
	}

	return &nsn
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
