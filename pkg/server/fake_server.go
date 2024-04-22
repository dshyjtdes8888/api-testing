package server

import (
	context "context"
	"net"

	"github.com/linuxsuren/api-testing/pkg/logging"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type fakeServer struct {
	UnimplementedRunnerServer
	version string
	err     error
}

var (
	fakeLogger = logging.DefaultLogger(logging.LogLevelInfo).WithName("fake_server")
)

// NewServer creates a fake server
func NewServer(version string, err error) RunnerServer {
	t := &fakeServer{
		version: version,
		err:     err,
	}
	return t
}

// Run runs the task
func (s *fakeServer) Run(ctx context.Context, in *TestTask) (*TestResult, error) {
	return &TestResult{}, s.err
}

// GetVersion returns the version
func (s *fakeServer) GetVersion(ctx context.Context, in *Empty) (reply *HelloReply, err error) {
	reply = &HelloReply{
		Message: s.version,
	}
	err = s.err
	return
}

// Sample returns a sample of the test task
func (s *fakeServer) Sample(ctx context.Context, in *Empty) (reply *HelloReply, err error) {
	reply = &HelloReply{
		Message: "",
	}
	err = s.err
	return
}

// NewFakeClient creates a fake client
func NewFakeClient(ctx context.Context, version string, err error) (RunnerClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	RegisterRunnerServer(baseServer, NewServer(version, err))
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			fakeLogger.Info("error serving server", "error", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fakeLogger.Info("error connecting to server", "error", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			fakeLogger.Info("error closing listener", "error", err)
		}
		baseServer.Stop()
	}

	client := NewRunnerClient(conn)
	return client, closer
}
