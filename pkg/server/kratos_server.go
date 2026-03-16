package server

import (
	"log/slog"
	"os"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	consulapi "github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"

	krtlogger "github.com/miladystack/miladystack/pkg/logger/klog/kratos"
	genericoptions "github.com/miladystack/miladystack/pkg/options"
)

type KratosAppConfig struct {
	ID        string
	Name      string
	Version   string
	Metadata  map[string]string
	Registrar registry.Registrar
}

type KratosServer struct {
	kapp *kratos.App
}

func NewKratosServer(cfg KratosAppConfig, servers ...transport.Server) (*KratosServer, error) {
	kapp := kratos.New(
		kratos.ID(cfg.ID+"."+cfg.Name),
		kratos.Name(cfg.Name),
		kratos.Version(cfg.Version),
		kratos.Metadata(cfg.Metadata),
		kratos.Registrar(cfg.Registrar),
		kratos.Logger(NewKratosLogger(cfg.ID, cfg.Name, cfg.Version)),
		kratos.Server(servers...),
	)

	return &KratosServer{
		kapp: kapp,
	}, nil
}

func NewKratosLogger(id, name, version string) log.Logger {
	return log.With(krtlogger.NewLogger(),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", name,
		"service.version", version,
	)
}

func (s *KratosServer) RunOrDie() {
	slog.Info("start to listening the incoming requests", "protocol", "kratos")
	if err := s.kapp.Run(); err != nil {
		slog.Error("failed to serve kratos application", "error", err)
		os.Exit(1)
	}
}

func (s *KratosServer) GracefulStop() {
	slog.Info("gracefully stop kratos application")
	if err := s.kapp.Stop(); err != nil {
		slog.Error("Failed to gracefully shutdown kratos application", "error", err)
	}
}

func NewEtcdRegistrar(opts *genericoptions.EtcdOptions) registry.Registrar {
	if opts == nil {
		panic("etcd registrar options must be set.")
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   opts.Endpoints,
		DialTimeout: opts.DialTimeout,
		TLS:         opts.TLSOptions.MustTLSConfig(),
		Username:    opts.Username,
		Password:    opts.Password,
	})
	if err != nil {
		panic(err)
	}
	r := etcd.New(client)
	return r
}

func NewConsulRegistrar(opts *genericoptions.ConsulOptions) registry.Registrar {
	if opts == nil {
		panic("consul registrar options must be set.")
	}

	c := consulapi.DefaultConfig()
	c.Address = opts.Addr
	c.Scheme = opts.Scheme
	cli, err := consulapi.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(false))
	return r
}
