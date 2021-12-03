package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"net"
	"os"
	"pipelines.kubeflow.org/events/eventsources/model_update/server"
	"pipelines.kubeflow.org/events/logging"

	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type eventingServer struct {
	server.UnimplementedEventingServer
}

func (es *eventingServer) StartEventSource(source *server.EventSource, stream server.Eventing_StartEventSourceServer) error {
	logger := logging.FromContext(context.Background())

	for true {
		event := &server.Event{
			Name: "MyEvent",
		}

		logger.Info("Sending event", "event", event)
		stream.Send(event)

		time.Sleep(1000 * 1000 * 1000)
	}

	return nil
}

func main() {
	var err error
	var logger *logr.Logger

	logger, err = logging.NewLogger()
	if err != nil {
		logger.Error(err, "failed to create logger")
		os.Exit(1)
	}

	logging.WithLogger(context.Background(), logger)

	flag.Parse()
	var lis net.Listener

	lis, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		logger.Error(err, "failed to listen")
		os.Exit(1)
	}

	s := grpc.NewServer()
	server.RegisterEventingServer(s, &eventingServer{})
	logger.Info(fmt.Sprintf("server listening at %s", lis.Addr()))
	if err := s.Serve(lis); err != nil {
		logger.Error(err, "failed to serve")
		os.Exit(1)
	}
}
