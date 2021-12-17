package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
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
	var k8sConfig *rest.Config
	var err error

	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(kubeconfigPath); err == nil {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", "")
	}

	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(k8sConfig)
}

type CmdArguments struct {
	Port              int
	MetadataStoreAddr string
	ZapLogLevel       zapcore.Level
}

func ParseCmdArguments() (CmdArguments, error) {
	port := flag.Int("port", 50051, "The server port")
	metadataStoreAddr := flag.String("mlmd-url", "", "The MLMD gRPC URL (required)")
	zapLogLevel := zap.LevelFlag("zap-log-level", zapcore.InfoLevel, "The Zap log level")
	flag.Parse()

	if *metadataStoreAddr == "" {
		return CmdArguments{}, fmt.Errorf("mlmd-url must be specified")
	}

	return CmdArguments{
		Port:              *port,
		MetadataStoreAddr: *metadataStoreAddr,
		ZapLogLevel:       *zapLogLevel,
	}, nil
}

func main() {
	cmdArguments, err := ParseCmdArguments()
	if err != nil {
		panic(fmt.Sprintf("failed to parse command line arguments: %v", err))
	}

	logger, err := logging.NewLogger(cmdArguments.ZapLogLevel)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	k8sClient, err := createK8sClient()
	if err != nil {
		logger.Error(err, "failed to create k8s client")
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", cmdArguments.Port))
	if err != nil {
		logger.Error(err, "failed to listen")
		os.Exit(1)
	}

	conn, err := grpc.Dial(cmdArguments.MetadataStoreAddr, grpc.WithInsecure())
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
