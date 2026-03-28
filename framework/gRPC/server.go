package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

type server struct {
	srv *grpc.Server
	cfg contracts.Config
	log contracts.Log
}

// NewServer 根据配置创建 gRPC Server 实例。
func NewServer(cfg contracts.Config, log contracts.Log) (contracts.GRPCServer, error) {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			RecoveryInterceptor(log),
			LoggingInterceptor(log),
		),
		grpc.ChainStreamInterceptor(
			StreamRecoveryInterceptor(log),
			StreamLoggingInterceptor(log),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:      time.Duration(cfg.GetInt("grpc.max_conn_age_sec", 300)) * time.Second,
			MaxConnectionAgeGrace: time.Duration(cfg.GetInt("grpc.max_conn_age_grace_sec", 5)) * time.Second,
			Time:                  time.Duration(cfg.GetInt("grpc.keepalive_time_sec", 60)) * time.Second,
			Timeout:               time.Duration(cfg.GetInt("grpc.keepalive_timeout_sec", 20)) * time.Second,
		}),
		grpc.MaxRecvMsgSize(cfg.GetInt("grpc.max_recv_msg_size_mb", 4) * 1024 * 1024),
		grpc.MaxSendMsgSize(cfg.GetInt("grpc.max_send_msg_size_mb", 4) * 1024 * 1024),
	}

	// 若配置了 TLS，加载证书
	// certFile := cfg.GetString("grpc.tls.cert_file", "")
	// keyFile := cfg.GetString("grpc.tls.key_file", "")
	// if certFile != "" && keyFile != "" {
	//     creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	//     if err != nil {
	//         return nil, fmt.Errorf("[GoFast/grpc] load tls: %w", err)
	//     }
	//     opts = append(opts, grpc.Creds(creds))
	// }

	srv := grpc.NewServer(opts...)

	// 开发模式开启反射（方便 grpcurl 调试）
	if cfg.GetString("grpc.mode", "debug") != "release" {
		reflection.Register(srv)
	}

	return &server{srv: srv, cfg: cfg, log: log}, nil
}

func (s *server) RegisterService(desc *grpc.ServiceDesc, impl any) {
	s.srv.RegisterService(desc, impl)
}

func (s *server) Run(addr ...string) error {
	address := fmt.Sprintf("%s:%d",
		s.cfg.GetString("grpc.host", "0.0.0.0"),
		s.cfg.GetInt("grpc.port", 9000))
	if len(addr) > 0 && addr[0] != "" {
		address = addr[0]
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("[GoFast/grpc] listen %s: %w", address, err)
	}
	s.log.Infof("[GoFast/grpc] server listening on %s", address)
	return s.srv.Serve(lis)
}

func (s *server) Shutdown() {
	s.log.Info("[GoFast/grpc] graceful stop...")
	s.srv.GracefulStop()
}

func (s *server) RawServer() *grpc.Server {
	return s.srv
}
