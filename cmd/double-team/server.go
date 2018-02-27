package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/msales/double-team"
	"github.com/msales/double-team/producer"
	"github.com/msales/double-team/server"
	"github.com/msales/double-team/server/middleware"
	"gopkg.in/urfave/cli.v1"
)

func runServer(c *cli.Context) {
	ctx, err := newContext(c)
	if err != nil {
		log.Fatal(err.Error())
	}

	kafkaProducer, err := newKafkaProducer(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	s3Producer, err := newS3Producer(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	app, err := newApplication(ctx, []producer.Producer{kafkaProducer, s3Producer})
	if err != nil {
		log.Fatal(err.Error())
	}

	port := c.String(FlagPort)
	srv := newServer(ctx, app)
	h := http.Server{Addr: ":" + port, Handler: srv}
	go func() {
		ctx.logger.Info(fmt.Sprintf("Starting server on port %s", port))
		if err := h.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}()

	quit := listenForSignals()
	<-quit

	// Close the server
	ctxServer, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h.Shutdown(ctxServer)
	ctx.logger.Info("Draining queues")

	// Close the application
	if err := app.Close(); err != nil {
		ctx.logger.Error(err.Error())
	}

	ctx.logger.Info("Server stopped gracefully")
}

func newServer(ctx *Context, app *double_team.Application) http.Handler {
	s := server.New(app)

	h := middleware.Common(s)
	return middleware.WithContext(ctx, h)
}

func listenForSignals() chan bool {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		done <- true
	}()

	return done
}
