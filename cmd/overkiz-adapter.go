package main

import (
	"context"
	"flag"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"overkiz-adapter/internal/config"
	"overkiz-adapter/internal/domain"
	"overkiz-adapter/internal/http"
	"overkiz-adapter/internal/log"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	syncGroup, syncGroupContext := errgroup.WithContext(ctx)

	// Parse command line parameters.
	configFile := *flag.String("config-file", "config.json", "Full path to the configuration file")
	flag.Parse()

	// Load configuration
	configuration, err := config.LoadConfiguration(configFile)
	if err != nil {
		log.Fatalf("Unable to load configuration file: %s", err.Error())
		syscall.Exit(-1)
	}

	overkiz, err := domain.NewOverkiz(configuration.Region, configuration.Pod, configuration.UserID, configuration.Password)
	if err != nil {
		log.Fatalf("Unable to connect to Overkiz: %s", err.Error())
		syscall.Exit(-1)
	}

	// Start the http server
	httpServer, err := http.NewServer(configuration.Http, overkiz)
	if err != nil {
		log.Warningf("Failed to create http server: %s", err.Error())
	}
	syncGroup.Go(func() error {
		return httpServer.Start()
	})
	syncGroup.Go(func() error {
		<-syncGroupContext.Done()
		if httpServer != nil {
			err = httpServer.Shutdown(context.Background())
			if err != nil {
				log.Warningf("Failed to stop http server: %s", err.Error())
			}
		}
		return nil
	})

	if err = syncGroup.Wait(); err != nil {
		log.Errorf("%v", err)
	}
}
