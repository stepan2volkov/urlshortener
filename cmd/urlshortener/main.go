package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"

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
	app := app.NewApp(store, logger)
	rt := router.NewRouter(app, logger)
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
