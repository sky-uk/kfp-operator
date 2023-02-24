package common

import (
	"encoding/json"
	"strings"
)

type NamespacedName struct {
	Name      string `json:"-"`
	Namespace string `json:"-"`
}

func (nsn NamespacedName) String() string {
	return strings.Join([]string{nsn.Name, nsn.Namespace}, "/")
}

func NamespacedNameFromString(namespacedName string) NamespacedName {
	splits := strings.Split(namespacedName, "/")

	if len(splits) < 2 {
		return NamespacedName{
			Name: namespacedName,
		}
	}

	return NamespacedName{
		Name:      splits[0],
		Namespace: splits[1],
	}
}

func (nsn *NamespacedName) MarshalJSON() ([]byte, error) {
	return json.Marshal(nsn.String())
}

func (nsn *NamespacedName) UnmarshalJSON(bytes []byte) error {
	var pidStr string
	err := json.Unmarshal(bytes, &pidStr)
	if err != nil {
		return err
	}

	*nsn = NamespacedNameFromString(pidStr)

	return nil
}
