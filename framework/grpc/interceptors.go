package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// ── Unary 拦截器 ──────────────────────────────────────────────────────

// RecoveryInterceptor Unary panic 恢复拦截器。
func RecoveryInterceptor(log contracts.Log) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("[GoFast/grpc] panic: %v\n%s", r, debug.Stack())
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// LoggingInterceptor Unary 请求日志拦截器。
func LoggingInterceptor(log contracts.Log) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(start)
		if err != nil {
			log.Errorf("[GoFast/grpc] %s | %v | %s", info.FullMethod, elapsed, err)
		} else {
			log.Infof("[GoFast/grpc] %s | %v", info.FullMethod, elapsed)
		}
		return resp, err
	}
}

// ── Stream 拦截器 ─────────────────────────────────────────────────────

// StreamRecoveryInterceptor Stream panic 恢复拦截器。
func StreamRecoveryInterceptor(log contracts.Log) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("[GoFast/grpc] stream panic: %v\n%s", r, debug.Stack())
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}

// StreamLoggingInterceptor Stream 请求日志拦截器。
func StreamLoggingInterceptor(log contracts.Log) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, ss)
		log.Infof("[GoFast/grpc] stream %s | %v | err=%v", info.FullMethod, time.Since(start), err)
		return err
	}
}
