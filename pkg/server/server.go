package server

import (
	"context"
	"net/http"
	"time"

	"k8s.io/klog/v2"
)

// Server 定义所有服务器类型的接口.
type Server interface {
	// RunOrDie 运行服务器，如果运行失败会退出程序（OrDie的含义所在）.
	RunOrDie()
	// GracefulStop 方法用来优雅关停服务器。关停服务器时需要处理 context 的超时时间.
	GracefulStop(ctx context.Context)
}

// Serve starts the server and blocks until the context is canceled.
// It ensures the server is gracefully shut down when the context is done.
func Serve(ctx context.Context, srv Server) error {
	go srv.RunOrDie()

	// Block until the context is canceled or terminated.
	<-ctx.Done()

	// Shutdown the server gracefully.
	klog.InfoS("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully stop the server.
	srv.GracefulStop(ctx)

	klog.InfoS("Server exited successfully.")

	return nil
}

// protocolName 从 http.Server 中获取服务使用的协议名称
//
// 参数说明：
//
//	server: http.Server 实例指针，允许传入 nil
//
// 返回值：
//
//	协议名称（http/https），默认返回 http
func protocolName(server *http.Server) string {
	if server == nil {
		return "http"
	}

	if server.TLSConfig != nil && len(server.TLSConfig.Certificates) > 0 {
		return "https"
	}

	return "http"
}
