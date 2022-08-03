package main

import (
	"flag"
	"fmt"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"net"
	"os"
	"path/filepath"
	"pipelines.kubeflow.org/events/eventsources/generic"
	"pipelines.kubeflow.org/events/eventsources/run_completion"
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
	KfpApiAddr        string
	ZapLogLevel       zapcore.Level
}

func ParseCmdArguments() (CmdArguments, error) {
	port := flag.Int("port", 50051, "The server port")
	metadataStoreAddr := flag.String("mlmd-url", "", "The MLMD gRPC URL (required)")
	kfpApiAddr := flag.String("kfp-url", "", "The KFP gRPC URL (required)")
	zapLogLevel := zap.LevelFlag("zap-log-level", zapcore.InfoLevel, "The Zap log level")
	flag.Parse()

	if *metadataStoreAddr == "" {
		return CmdArguments{}, fmt.Errorf("mlmd-url must be specified")
	}

	if *kfpApiAddr == "" {
		return CmdArguments{}, fmt.Errorf("kfp-url must be specified")
	}

	return CmdArguments{
		Port:              *port,
		MetadataStoreAddr: *metadataStoreAddr,
		KfpApiAddr:        *kfpApiAddr,
		ZapLogLevel:       *zapLogLevel,
	}, nil
}

func ConnectToMetadataStore(address string) (*run_completion.GrpcMetadataStore, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &run_completion.GrpcMetadataStore{
		MetadataStoreServiceClient: ml_metadata.NewMetadataStoreServiceClient(conn),
	}, nil
}

func ConnectToKfpApi(address string) (*run_completion.GrpcKfpApi, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &run_completion.GrpcKfpApi{
		RunServiceClient: go_client.NewRunServiceClient(conn),
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

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cmdArguments.Port))
	if err != nil {
		logger.Error(err, "failed to listen")
		os.Exit(1)
	}

	metadataStore, err := ConnectToMetadataStore(cmdArguments.MetadataStoreAddr)
	if err != nil {
		logger.Error(err, "failed to connect to metadata store")
		os.Exit(1)
	}

	kfpApi, err := ConnectToKfpApi(cmdArguments.KfpApiAddr)
	if err != nil {
		logger.Error(err, "failed to connect to KFP API")
		os.Exit(1)
	}

	s := grpc.NewServer()
	generic.RegisterEventingServer(s, &run_completion.EventingServer{
		K8sClient:     k8sClient,
		Logger:        logger,
		MetadataStore: metadataStore,
		KfpApi:        kfpApi,
	})

	logger.Info(fmt.Sprintf("server listening at %s", lis.Addr()))
	if err := s.Serve(lis); err != nil {
		logger.Error(err, "failed to serve")
		os.Exit(1)
	}
}
