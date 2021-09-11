package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/stepan2volkov/urlshortener/api/router"
	"github.com/stepan2volkov/urlshortener/api/server"
	"github.com/stepan2volkov/urlshortener/app"
	"github.com/stepan2volkov/urlshortener/app/config"
	"github.com/stepan2volkov/urlshortener/db/memstore"
	"github.com/stepan2volkov/urlshortener/db/pgstore"
)

var configPath string

func main() {
	log.Println("Build Commit:", config.BuildCommit)
	log.Println("Build Time:", config.BuildTime)
	flag.StringVar(&configPath, "config", "config.yaml", "path to config")
	flag.Parse()
	conf, err := config.GetConfig(configPath)
	if err != nil {
		log.Fatalf("error parsing config: %v\n", err)
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	var store app.URLStore
	switch {
	case conf.DSN == "memory":
		store = memstore.NewMemStore()
	case strings.HasPrefix(conf.DSN, "postgres://"):
		store, err = pgstore.NewPgStore(conf.DSN)
		if err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalf("unknown store value in config: \"%v\"\n", conf.DSN)
	}
	app := app.NewApp(store)
	rt := router.NewRouter(app)
	srv := server.NewServer(conf, rt)

	srv.Start()
	<-ctx.Done()
	srv.Stop()
}
