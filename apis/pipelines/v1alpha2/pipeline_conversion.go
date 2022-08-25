package v1alpha2

import "sigs.k8s.io/controller-runtime/pkg/conversion"

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	return nil
}
