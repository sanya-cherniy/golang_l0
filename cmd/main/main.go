package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"

	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"

	"l0/internal/config"
	"l0/internal/order"
	"l0/pkg/client/broker"
	"l0/pkg/client/postgres"
	"l0/pkg/logging"
)

const (
	logFile            = "logs/server.log"
	serverStartTimeout = 10 * time.Second
)

func main() {
	logging.InitLogger(logFile)
	logger, err := logging.GetLogger(logFile)
	if err != nil {
		panic(err)
	}
	logger.Info("create router")

	router := gin.Default()
	cfg := config.GetConfig(logFile)

	start(router, cfg, logger)
}

func start(router *gin.Engine, cfg *config.Config, logger *logging.Logger) {
	logger.Info("start application")

	var listener net.Listener
	var listenErr error

	if cfg.Listen.Type == "socket" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("create socket")
		socketPath := path.Join(appDir, "app.sock")
		logger.Debugf("socket path: %s", socketPath)

		logger.Info("listen unix socket")
		listener, listenErr = net.Listen("unix", socketPath)
		logger.Infof("server is listening on unix socket: %s", socketPath)

	} else {
		logger.Info("listen port")
		listener, listenErr = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		logger.Infof("server is listening on %s:%s", cfg.Listen.BindIP, cfg.Listen.Port)
	}
	if listenErr != nil {
		logger.Fatal(listenErr)
	}

	memoryStorage := order.NewMemoryStorage()
	pool, err := postgres.NewClient(context.Background(), cfg.Storage)

	if err != nil {
		logger.Fatal(err)
	}
	brokerProducer, err := broker.NewProducer(cfg.Brokers)

	if err != nil {
		logger.Fatal(err)
	}
	go memoryStorage.CashChecker(cfg.LifeTime)

	handler := order.NewHandler(memoryStorage, logger, pool, brokerProducer)
	handler.Register(router)

	func(ctx context.Context) {
		ctx, cancel := context.WithCancel(ctx)
		go func() {
			defer cancel()
			logger.Fatal(router.RunListener(listener))
		}()

		notifyCtx, notify := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		defer notify()

		go func() {
			defer cancel()
			<-notifyCtx.Done()
			closer := make(chan struct{})

			go func() {
				closer <- struct{}{}
			}()

			shutdownCtx, shutdown := context.WithTimeout(context.Background(), serverStartTimeout)
			defer shutdown()
			runtime.Gosched()

			select {
			case <-closer:
				logger.Info("shutting down gracefully")
			case <-shutdownCtx.Done():
				logger.Info("shutting down forcefully")
			}
		}()

		<-ctx.Done()
		cancel()

	}(context.Background())

}
