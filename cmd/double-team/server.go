package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/msales/double-team/streaming"
	"github.com/msales/pkg/v3/clix"
	"github.com/msales/pkg/v3/log"
	"github.com/msales/pkg/v3/stats"
	"gopkg.in/urfave/cli.v1"
)

func runServer(c *cli.Context) {
	ctx, err := clix.NewContext(c)
	if err != nil {
		log.Fatal(ctx, err.Error())
	}

	go stats.RuntimeFromContext(ctx, 10*time.Second)

	kafkaProducer, err := newKafkaProducer(ctx)
	if err != nil {
		log.Fatal(ctx, err.Error())
	}

	s3Producer, err := newS3Producer(ctx)
	if err != nil {
		log.Fatal(ctx, err.Error())
	}

	app, err := newApplication(ctx, []streaming.Producer{kafkaProducer, s3Producer}, c.Int(FlagQueueSize))
	if err != nil {
		log.Fatal(ctx, err.Error())
	}

	port := c.String(clix.FlagPort)
	srv := newServer(ctx, app)
	h := http.Server{Addr: ":" + port, Handler: srv}
	log.Info(ctx, fmt.Sprintf("Starting server on port %s", port))
	go func() {
		if err := h.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(ctx, err)
			}
		}
	}()

	<-clix.WaitForSignals()

	// Close the server
	ctxServer, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.Shutdown(ctxServer); err != nil {
		log.Error(ctx, err.Error())
	}
	log.Info(ctx, "Draining queues")

	// Close the application
	if err := app.Close(); err != nil {
		log.Error(ctx, err.Error())
	}

	log.Info(ctx, "Server stopped gracefully")
}
