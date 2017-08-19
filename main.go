package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boltdb/bolt"

	"go.uber.org/zap"
)

const (
	DatabasePath   string        = "database/showcase.db"
	DefaultTimeout time.Duration = 15 * time.Second
)

func main() {
	httpAddr := flag.String("http.addr", ":8080", "HTTP listen address")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	bolt, err := bolt.Open(DatabasePath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer bolt.Close()

	var (
		ctx   = context.TODO()
		cache = NewDiskCache(DefaultExpiration, DefaultEvictionInterval)
		db    = NewCachedDatabase(bolt, cache, logger)
	)

	c, err := db.Setup(ctx)
	if err != nil {
		panic(err)
	}

	var (
		composer     = NewComposer(db, logger)
		renderer     = NewRenderer()
		manager      = NewMediaManager(cache)
		builder      = NewSitemapBuiler(db, logger, SitemapInterval)
		configurator = NewConfigurator(composer, renderer, manager, builder)
		finalizer    = NewFinalizer(cache, builder)
	)
	defer finalizer.Finalize()

	var (
		server = NewServer(configurator, db, cache, composer, renderer, manager, logger)
		c0, c1 = configurator.Configure(c)
	)

	go func() {
		for {
			select {
			case <-c0:
				return
			case e := <-c1:
				panic(e)
			case <-time.After(ConfigurationTimeoutInterval):
				panic(ErrConfigurationTimedOut)
			}
		}
	}()

	<-c0

	go builder.Run()
	go manager.PopulateEtagCache()

	var (
		h = NewSwappableServeMux(http.NewServeMux())
		s = &http.Server{
			Handler:      h,
			Addr:         *httpAddr,
			WriteTimeout: DefaultTimeout,
			ReadTimeout:  DefaultTimeout,
		}
	)
	defer s.Close()

	errs := make(chan error, 2)
	go func() {
		logger.Info("app", zap.String("transport", "HTTP"), zap.String("addr", *httpAddr))
		errs <- s.ListenAndServe()
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		if !c.SetupCompleted {
			r, d, n := server.NewSetupRouter()

			h.Handle("/", r)

			logger.Info("app", zap.String("event", "entering setup mode"))

			if !<-d {
				errs <- fmt.Errorf("%s", "could not perform setup")
				return
			}

			h.Swap(http.NewServeMux())

			h.Handle("/admin/", server.NewAdminRouter())
			h.Handle("/", server.NewRouter())

			n <- true
		} else {
			h.Handle("/admin/", server.NewAdminRouter())
			h.Handle("/", server.NewRouter())
		}
	}()

	logger.Warn("app", zap.String("event", "terminating"), zap.Error(<-errs))
}
