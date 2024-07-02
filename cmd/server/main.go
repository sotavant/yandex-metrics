package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/handlers"
	"github.com/sotavant/yandex-metrics/internal/server/middleware"
	"github.com/sotavant/yandex-metrics/internal/utils"
	pb "github.com/sotavant/yandex-metrics/proto"
	"google.golang.org/grpc"
)

// Build info.
// Need define throw ldflags:
//
//	go build -ldflags "-X main.buildVersion=0.1 -X 'main.buildDate=$(date +'%Y/%m/%d')' -X 'main.buildCommit=$(git rev-parse --short HEAD)'"
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	var listen net.Listener
	var srv http.Server
	var s *grpc.Server
	internal.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	ctx := context.Background()
	internal.InitLogger()

	appInstance, err := server.InitApp(ctx)
	if err != nil {
		panic(err)
	}

	if appInstance.Config.UseGRPC {
		listen, err = net.Listen("tcp", appInstance.Config.Addr)
		if err != nil {
			internal.Logger.Fatalw("failed to listen", "err", err)
		}

		s = initGRPCServer(appInstance)
	} else {
		r := initRouters(appInstance)
		srv = http.Server{Addr: appInstance.Config.Addr, Handler: r}
	}

	jobsDone := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigint
		if appInstance.Config.UseGRPC {
			s.GracefulStop()
		} else {
			if err = srv.Shutdown(ctx); err != nil {
				internal.Logger.Infow("shutdown err", "err", err)
			}
		}

		appInstance.SyncFs(ctx)
		close(jobsDone)
		internal.Logger.Infow("shutdown complete")
	}()

	go func() {
		if appInstance.Config.UseGRPC {
			if err = s.Serve(listen); err != nil {
				internal.Logger.Fatalw("failed to grpc serve", "err", err)
			}
		} else {
			if err = srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
				internal.Logger.Infow("http server err", "err", err)
			}
		}
	}()

	go func() {
		if appInstance.Fs == nil {
			return
		}
		if err = appInstance.Fs.SyncByInterval(ctx, appInstance.Storage); err != nil {
			panic(err)
		}
	}()

	<-jobsDone
}

func initRouters(app *server.App) *chi.Mux {

	r := chi.NewRouter()

	hasher := middleware.NewHasher(app.Config.HashKey)
	crypto, err := middleware.NewCrypto(app.Config.CryptoKeyPath)
	ipChecker := middleware.NewIPChecker(app.Config.TrustedSubnet)

	if err != nil {
		internal.Logger.Fatalw("crypto initialization failed", "error", err)
	}

	if ipChecker != nil {
		r.Use(ipChecker.CheckIP)
	}

	r.Use(crypto.Handler)
	r.Use(hasher.Handler)
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithLogging)

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandler(app))
	r.Get("/value/{type}/{name}", handlers.GetValueHandler(app))
	r.Post("/update/", handlers.UpdateJSONHandler(app))
	r.Post("/updates/", handlers.UpdateBatchJSONHandler(app))
	r.Post("/value/", handlers.GetValueJSONHandler(app))
	r.Get("/", handlers.GetValuesHandler(app))
	r.Get("/ping", handlers.PingDBHandler(app.DBConn))

	initProfiling(r)

	return r
}

func initProfiling(r *chi.Mux) {
	r.HandleFunc("/pprof/*", pprof.Index)
	r.Handle("/pprof/heap", pprof.Handler("heap"))
}

func initGRPCServer(app *server.App) *grpc.Server {
	var ch *utils.Cipher
	var err error
	ch, err = utils.NewCipher(app.Config.CryptoKeyPath, "", app.Config.CryptoCertPath)
	if err != nil {
		internal.Logger.Fatalw("error initializing cipher", "err", err)
	}

	s := grpc.NewServer(grpc.Creds(ch.GetServerGRPCTransportCreds()))
	pb.RegisterMetricsServer(s, &MetricsServer{})

	return s
}
