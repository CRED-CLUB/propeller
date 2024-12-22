package grpcserver

import (
	"context"
	"net"
	"time"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
    "google.golang.org/grpc/reflection"
)

// GrpcServer holds grpc server
type GrpcServer struct {
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
	config             Config
	Server             *grpc.Server
}

// NewGrpc returns a GrpcServer
func NewGrpc(config Config) *GrpcServer {
	return &GrpcServer{
		streamInterceptors: nil,
		unaryInterceptors:  nil,
		config:             config,
	}
}

// WithUnaryInterceptors sets unary interceptors
func (s *GrpcServer) WithUnaryInterceptors(unaryInterceptors ...grpc.UnaryServerInterceptor) *GrpcServer {
	s.unaryInterceptors = unaryInterceptors
	return s
}

// WithStreamInterceptors sets stream interceptors
func (s *GrpcServer) WithStreamInterceptors(streamInterceptors ...grpc.StreamServerInterceptor) *GrpcServer {
	s.streamInterceptors = streamInterceptors
	return s
}

type registerGrpcHandlers func(server *grpc.Server) error

// InitGRPCServer with handlers and interceptors
func (s *GrpcServer) InitGRPCServer(
	ctx context.Context,
	registerGrpcHandlers registerGrpcHandlers) error {

	grpcServer, err := s.newGrpcServer(registerGrpcHandlers)
	if err != nil {
		return err
	}

	s.Server = grpcServer
	return nil
}

// Run the grpc server
func (s *GrpcServer) Run(ctx context.Context) error {

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}

	// wait for ctx.Done() in a goroutine and stop the server gracefully
	go func() {
		<-ctx.Done()
		s.Server.GracefulStop()
	}()

	// Start gRPC server
	return s.Server.Serve(listener)
}

func (s *GrpcServer) newGrpcServer(r registerGrpcHandlers) (*grpc.Server, error) {

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(s.streamInterceptors...)),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(s.unaryInterceptors...)),
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Time:    time.Duration(s.config.PingIntervalInSec) * time.Second,
				Timeout: time.Duration(s.config.PingResponseTimeoutInSec) * time.Second,
			},
		),
	)

    reflection.Register(grpcServer)

	err := r(grpcServer)
	if err != nil {
		return nil, err
	}

	return grpcServer, nil
}
