package v1alpha2

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"sort"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Pipeline)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status
	dst.Spec.Env = mapToNamedValues(src.Spec.Env)
	dst.Spec.BeamArgs = mapToNamedValues(src.Spec.BeamArgs)
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Pipeline)

	var err error

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status
	dst.Spec.Env, err = namedValuesToMap(src.Spec.Env)
	if err != nil {
		return err
	}
	dst.Spec.BeamArgs, err = namedValuesToMap(src.Spec.BeamArgs)
	if err != nil {
		return err
	}
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents
	return nil
}

func namedValuesToMap(namedValues []apis.NamedValue) (map[string]string, error) {
	if len(namedValues) == 0 {
		return nil, nil
	}

	values := make(map[string]string, len(namedValues))

	for _, nv := range namedValues {
		if _, exists := values[nv.Name]; exists {
			return nil, fmt.Errorf("duplicate entry: %s", nv.Name)
		}

		values[nv.Name] = nv.Value
	}

	return values, nil
}

func mapToNamedValues(values map[string]string) []apis.NamedValue {
	var namedValues []apis.NamedValue

	for k, v := range values {
		namedValues = append(namedValues, apis.NamedValue{
			Name: k, Value: v,
		})
	}

	sort.Slice(namedValues, func(i, j int) bool {
		if namedValues[i].Name != namedValues[j].Name {
			return namedValues[i].Name < namedValues[j].Name
		} else {
			return namedValues[i].Value < namedValues[j].Value
		}
	})

	return namedValues
}
