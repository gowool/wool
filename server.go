package wool

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gowool/wool/internal"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var StopSignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGTSTP,
}

type ServerConfig struct {
	DisableHTTP2      bool          `mapstructure:"disable_http2"`
	HidePort          bool          `mapstructure:"hide_port"`
	CertFile          string        `mapstructure:"cert_file"`
	KeyFile           string        `mapstructure:"key_file"`
	Address           string        `mapstructure:"address"`
	Network           string        `mapstructure:"network"`
	MaxHeaderBytes    int           `mapstructure:"max_header_bytes"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	GracefulTimeout   time.Duration `mapstructure:"graceful_timeout"`
}

func (cfg *ServerConfig) Init() {
	if cfg.Network == "" {
		cfg.Network = "tcp"
	}
	if cfg.Address == "" {
		cfg.Address = ":0"
	}
	if cfg.GracefulTimeout == 0 {
		cfg.GracefulTimeout = 10 * time.Second
	}
}

func (cfg *ServerConfig) Server(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

type Server struct {
	cfg             *ServerConfig
	server          *http.Server
	listener        net.Listener
	Log             *slog.Logger
	CertFilesystem  fs.FS
	TLSConfig       func(tlsConfig *tls.Config)
	ListenerAddr    func(addr net.Addr)
	BeforeServe     func(s *http.Server) error
	OnShutdownError func(err error)
}

func NewServer(cfg *ServerConfig) *Server {
	cfg.Init()

	return &Server{cfg: cfg, Log: Logger().WithGroup("server")}
}

func (s *Server) StartC(ctx context.Context, handler http.Handler) error {
	if err := s.init(handler); err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(ctx, StopSignals...)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(s.serve)

	g.Go(func() error {
		return s.gracefulShutdown(ctx)
	})

	s.Log.Debug("press Ctrl+C to stop")

	return g.Wait()
}

func (s *Server) Start(handler http.Handler) error {
	if err := s.init(handler); err != nil {
		return err
	}

	return s.serve()
}

func (s *Server) Close() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			if s.OnShutdownError != nil {
				s.OnShutdownError(err)
				return nil
			}
			s.Log.Error("failed to shutdown server within given timeout", err)
			return err
		}
	}
	return nil
}

func (s *Server) init(handler http.Handler) error {
	listener, err := s.createListener()
	if err != nil {
		return err
	}

	s.listener = listener
	s.server = s.cfg.Server(handler)

	s.Log.Info("http(s) server starting")

	if !s.cfg.HidePort {
		s.Log.Info(fmt.Sprintf("http(s) server started on %s", listener.Addr()))
	}

	if s.BeforeServe != nil {
		if err = s.BeforeServe(s.server); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) serve() error {
	if err := s.server.Serve(s.listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) createListener() (net.Listener, error) {
	var tlsConfig *tls.Config = nil

	if s.cfg.CertFile != "" && s.cfg.KeyFile != "" {
		certFs := s.CertFilesystem
		if certFs == nil {
			certFs = os.DirFS(".")
		}

		cert, err := fileContent(s.cfg.CertFile, certFs)
		if err != nil {
			return nil, err
		}
		key, err := fileContent(s.cfg.KeyFile, certFs)
		if err != nil {
			return nil, err
		}
		cer, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
		if !s.cfg.DisableHTTP2 {
			tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2")
		}
	}

	if s.TLSConfig != nil {
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		s.TLSConfig(tlsConfig)
	}

	var listener net.Listener
	var err error
	if tlsConfig != nil {
		listener, err = tls.Listen(s.cfg.Network, s.cfg.Address, tlsConfig)
	} else {
		listener, err = net.Listen(s.cfg.Network, s.cfg.Address)
	}
	if err != nil {
		return nil, err
	}

	if s.ListenerAddr != nil {
		s.ListenerAddr(listener.Addr())
	}

	return listener, nil
}

func (s *Server) gracefulShutdown(ctx context.Context) error {
	<-ctx.Done()

	forceCtx, cancel1 := signal.NotifyContext(context.Background(), StopSignals...)
	defer cancel1()

	forceCtx, cancel2 := context.WithTimeout(forceCtx, s.cfg.GracefulTimeout)
	defer cancel2()

	s.Log.Info("http(s) server stopping")
	s.Log.Debug("press Ctrl+C to force stopping")

	defer s.Log.Info("http(s) server stopped")

	return s.Shutdown(forceCtx)
}

func fileContent(cert string, certFilesystem fs.FS) (content []byte, err error) {
	content, err = fs.ReadFile(certFilesystem, cert)
	if os.IsNotExist(err) {
		return internal.StringToBytes(cert), nil
	}
	return
}
