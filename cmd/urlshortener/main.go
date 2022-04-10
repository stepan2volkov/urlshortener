package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/opentracing/opentracing-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/stepan2volkov/urlshortener/api/router"
	"github.com/stepan2volkov/urlshortener/api/server"
	"github.com/stepan2volkov/urlshortener/app"
	"github.com/stepan2volkov/urlshortener/app/config"
	"github.com/stepan2volkov/urlshortener/db/memstore"
	"github.com/stepan2volkov/urlshortener/db/pgstore"
)

var configPath string

func main() {
	logger := getLogger()
	// Information about current build
	logger.Info("app started",
		zap.String("Build Commit", config.BuildCommit),
		zap.String("Build Time", config.BuildTime),
	)

	tracer, closer := getTracer("example", logger)
	defer closer.Close()

	// Getting configuration
	flag.StringVar(&configPath, "config", "", "path to config")
	flag.Parse()
	conf, err := config.GetConfig(configPath)
	if err != nil {
		logger.Fatal("error when parsing config", zap.Error(err))
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	var store app.URLStore
	switch {
	case conf.DSN == "memory":
		store = memstore.NewMemStore()
	case strings.HasPrefix(conf.DSN, "postgres://"):
		store, err = pgstore.NewPgStore(conf.DSN)
		if err != nil {
			logger.Fatal("error when init pg storage", zap.Error(err))
		}
	default:
		logger.Fatal("unknown store value in config", zap.String("dsn", conf.DSN))
	}

	// Initialization and running application
	app := app.NewApp(store, logger, tracer)
	rt := router.NewRouter(app, logger, tracer)
	srv := server.NewServer(conf, rt, logger)
	srv.Start()

	<-ctx.Done()
	srv.Stop()
}

func getLogger() *zap.Logger {
	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl <= zapcore.ErrorLevel
	})
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		TimeKey:     "timestamp",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	})

	sync := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, sync, priority),
	)

	return zap.New(core)
}

func getTracer(service string, logger *zap.Logger) (opentracing.Tracer, io.Closer) {
	cfg := &jaegerConfig.Configuration{
		ServiceName: service,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "jaeger:6831",
		},
	}
	tracer, closer, err := cfg.NewTracer(jaegerConfig.Logger(&zapWrapper{logger: logger}))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

type zapWrapper struct {
	logger *zap.Logger
}

// Error logs a message at error priority
func (w *zapWrapper) Error(msg string) {
	w.logger.Error(msg)
}

// Infof logs a message at info priority
func (w *zapWrapper) Infof(msg string, args ...interface{}) {
	w.logger.Sugar().Infof(msg, args...)
}
