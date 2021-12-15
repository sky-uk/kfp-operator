package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"net"
	"os"
	"path/filepath"
	"pipelines.kubeflow.org/events/eventsources/generic"
	"pipelines.kubeflow.org/events/eventsources/model_update"
	"pipelines.kubeflow.org/events/logging"
	"pipelines.kubeflow.org/events/ml_metadata"
)

func createK8sClient() (dynamic.Interface, error) {
	var kubeconfigPath string

	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)

	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(k8sConfig)
}

func main() {
	port := flag.Int("port", 50051, "The server port")
	metadataStoreAddr := flag.String("mlmd-url", "", "The MLMD gRPC URL (required)")
	flag.Parse()

	logger, err := logging.NewLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	if *metadataStoreAddr == "" {
		logger.Info("mlmd-url must be specified")
		os.Exit(1)
	}

	k8sClient, err := createK8sClient()
	if err != nil {
		logger.Error(err, "failed to create k8s client")
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		logger.Error(err, "failed to listen")
		os.Exit(1)
	}

	conn, err := grpc.Dial(*metadataStoreAddr, grpc.WithInsecure())
	if err != nil {
		logger.Error(err, "failed to connect connect")
	}

	metadataStoreClient := ml_metadata.NewMetadataStoreServiceClient(conn)

	s := grpc.NewServer()
	generic.RegisterEventingServer(s, &model_update.EventingServer{
		K8sClient: k8sClient,
		Logger:    logger,
		MetadataStore: &model_update.GrpcMetadataStore{
			MetadataStoreServiceClient: metadataStoreClient,
		},
	})
	logger.Info(fmt.Sprintf("server listening at %s", lis.Addr()))
	if err := s.Serve(lis); err != nil {
		logger.Error(err, "failed to serve")
		os.Exit(1)
	}
}
