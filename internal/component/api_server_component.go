package component

import (
	"context"
	"fmt"
	"net/http"

	"github.com/CRED-CLUB/propeller/internal/component/apiserver"
	"github.com/CRED-CLUB/propeller/internal/config"
	"github.com/CRED-CLUB/propeller/internal/grpcserver"
	kvpkg "github.com/CRED-CLUB/propeller/internal/kv"
	"github.com/CRED-CLUB/propeller/internal/perror"
	pubsubpkg "github.com/CRED-CLUB/propeller/internal/pubsub"
	"github.com/CRED-CLUB/propeller/internal/push"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	grpctrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/google.golang.org/grpc"
)

// NewAPIServer returns a component instance
func NewAPIServer(config config.Config) (IComponent, error) {
	return &APIServer{config: config}, nil
}

// APIServer ...
type APIServer struct {
	config config.Config
}

// Start the API server
func (web *APIServer) Start(ctx context.Context) error {

	// Setup prom metrics
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	prometheus.MustRegister(srvMetrics)

	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		return nil
	}

	// initiates an error group
	grp, gctx := errgroup.WithContext(ctx)

	pubsub, err := pubsubpkg.New(ctx, web.config.Broker)
	if err != nil {
		return err
	}

	kv, err := kvpkg.New(ctx, web.config.Broker)
	if err != nil {
		return err
	}
	pushService := push.NewService(pubsub, kv, web.config)
	cancelCtx, cancelFunc := context.WithCancel(gctx)
	pushGrpcService := apiserver.NewPushServer(pushService, web.config)

	streamServerInterceptor := grpctrace.StreamServerInterceptor(
		grpctrace.WithServiceName(web.config.ServiceName),
		grpctrace.WithStreamMessages(false),
	)

	unaryServerInterceptor := grpctrace.UnaryServerInterceptor(
		grpctrace.WithServiceName(web.config.ServiceName),
		grpctrace.WithStreamMessages(true),
	)

	s := grpcserver.NewGrpc(web.config.Grpc).
		WithUnaryInterceptors(
			unaryServerInterceptor,
			srvMetrics.UnaryServerInterceptor(
				grpcprom.WithExemplarFromContext(exemplarFromContext))).
		WithStreamInterceptors(
			streamServerInterceptor,
			srvMetrics.StreamServerInterceptor(
				grpcprom.WithExemplarFromContext(exemplarFromContext)))
	err = s.InitGRPCServer(
		gctx,
		func(server *grpc.Server) error {
			pushv1.RegisterPushServiceServer(server, pushGrpcService)
			return nil
		},
	)
	if err != nil {
		logger.Ctx(ctx).Error(err.Error())
		cancelFunc()
		return err
	}
	srvMetrics.InitializeMetrics(s.Server)

	grp.Go(func() error {
		err := s.Run(gctx)
		return err
	})
	httpSrv := &http.Server{Addr: fmt.Sprintf(":%s", (web.config.HTTP.Port))}
	ws := apiserver.WebSocketServerWrapper{PushServer: pushGrpcService, Ctx: cancelCtx, CancelFunc: cancelFunc}
	grp.Go(func() error {
		m := http.NewServeMux()
		// Create HTTP handler for Prometheus metrics.
		m.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				// Opt into OpenMetrics e.g. to support exemplars.
				EnableOpenMetrics: true,
			},
		))
		m.HandleFunc(
			"/ws/connect",
			ws.WebSocketConnect,
		)

		httpSrv.Handler = m
		return httpSrv.ListenAndServe()
	})
	err = grp.Wait()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in starting component %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}
