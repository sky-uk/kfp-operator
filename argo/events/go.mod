module pipelines.kubeflow.org/events

go 1.16

require (
	github.com/argoproj/argo-workflows/v3 v3.1.8
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/mock v1.6.0
	github.com/onsi/gomega v1.17.0
	go.uber.org/zap v1.19.0
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	sigs.k8s.io/controller-runtime v0.7.0
)

require (
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/prometheus/client_golang v1.11.0 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/oauth2 v0.0.0-20210805134026-6f1e6394065a // indirect
	golang.org/x/tools v0.1.2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	k8s.io/apiextensions-apiserver v0.21.1 // indirect
)
