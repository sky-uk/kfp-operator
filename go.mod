module github.com/sky-uk/kfp-operator

go 1.26.4

require (
	cloud.google.com/go/aiplatform v1.126.0
	cloud.google.com/go/pubsub/v2 v2.6.1
	cloud.google.com/go/storage v1.64.0
	github.com/Masterminds/semver/v3 v3.5.0
	github.com/argoproj/argo-workflows/v3 v3.7.13
	github.com/distribution/reference v0.6.0
	github.com/evanphx/json-patch/v5 v5.9.11
	github.com/fsouza/fake-gcs-server v1.55.1
	github.com/go-chi/chi/v5 v5.3.1
	github.com/go-logr/logr v1.4.4
	github.com/go-logr/zapr v1.3.0
	github.com/go-openapi/runtime v0.32.6
	github.com/go-openapi/strfmt v0.27.0
	github.com/go-resty/resty/v2 v2.17.2
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/google/go-cmp v0.7.0
	github.com/googleapis/gax-go/v2 v2.23.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.1.0
	github.com/hashicorp/go-bexpr v0.1.16
	github.com/jarcoal/httpmock v1.4.1
	github.com/kubeflow/pipelines v0.0.0-20250902140602-fc3214d4b4d3
	github.com/nats-io/nats.go v1.52.0
	github.com/onsi/ginkgo/v2 v2.32.0
	github.com/onsi/gomega v1.42.1
	github.com/prometheus/client_golang v1.24.0
	github.com/prometheus/client_model v0.6.2
	github.com/samber/lo v1.53.0
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/thanhpk/randstr v1.0.6
	github.com/tidwall/sjson v1.2.5
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0
	go.opentelemetry.io/otel v1.44.0
	go.opentelemetry.io/otel/exporters/prometheus v0.66.0
	go.opentelemetry.io/otel/metric v1.44.0
	go.opentelemetry.io/otel/sdk v1.44.0
	go.opentelemetry.io/otel/sdk/metric v1.44.0
	go.uber.org/zap v1.28.0
	golang.org/x/sync v0.22.0
	google.golang.org/api v0.290.0
	google.golang.org/grpc v1.82.1
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af
	k8s.io/api v0.36.3
	k8s.io/apiextensions-apiserver v0.36.3
	k8s.io/apimachinery v0.36.3
	k8s.io/client-go v0.36.3
	k8s.io/utils v0.0.0-20260319190234-28399d86e0b5
	sigs.k8s.io/controller-runtime v0.24.1
)

require (
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.11.0 // indirect
	cloud.google.com/go/longrunning v1.2.0 // indirect
	cloud.google.com/go/monitoring v1.29.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.32.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.57.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.57.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20260202195803-dba9d589def2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.37.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.25.5 // indirect
	github.com/go-openapi/errors v0.22.8 // indirect
	github.com/go-openapi/jsonpointer v1.0.0 // indirect
	github.com/go-openapi/jsonreference v1.0.0 // indirect
	github.com/go-openapi/loads v0.25.0 // indirect
	github.com/go-openapi/runtime/server-middleware v0.30.0 // indirect
	github.com/go-openapi/spec v0.22.9 // indirect
	github.com/go-openapi/swag v0.25.5 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.5 // indirect
	github.com/go-openapi/swag/conv v0.27.3 // indirect
	github.com/go-openapi/swag/fileutils v0.27.3 // indirect
	github.com/go-openapi/swag/jsonname v0.26.1 // indirect
	github.com/go-openapi/swag/jsonutils v0.27.3 // indirect
	github.com/go-openapi/swag/loading v0.27.3 // indirect
	github.com/go-openapi/swag/mangling v0.27.3 // indirect
	github.com/go-openapi/swag/netutils v0.25.5 // indirect
	github.com/go-openapi/swag/pools v0.27.3 // indirect
	github.com/go-openapi/swag/stringutils v0.27.3 // indirect
	github.com/go-openapi/swag/typeutils v0.27.3 // indirect
	github.com/go-openapi/swag/yamlutils v0.27.3 // indirect
	github.com/go-openapi/validate v0.26.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	github.com/google/renameio/v2 v2.0.2 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.18 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.19.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/pointerstructure v1.2.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pelletier/go-toml/v2 v2.3.0 // indirect
	github.com/pkg/xattr v0.4.12 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/common v0.70.0 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.43.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.68.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/genproto v0.0.0-20260519071638-aa98bba5eb94 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260630182238-925bb5da69e7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260330154417-16be699c7b31 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.3 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace github.com/kubeflow/pipelines => github.com/kubeflow/pipelines v0.0.0-20250902140602-fc3214d4b4d3

tool golang.org/x/tools/cmd/stringer
