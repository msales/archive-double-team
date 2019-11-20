package main

import (
	"net/http"

	"github.com/msales/double-team"
	"github.com/msales/double-team/server"
	"github.com/msales/double-team/server/middleware"
	"github.com/msales/double-team/streaming"
	"github.com/msales/pkg/v3/clix"
)

// Server =============================

func newServer(ctx *clix.Context, app *doubleteam.Application) http.Handler {
	s := server.New(app)

	h := middleware.Common(s)
	return middleware.WithContext(ctx, h)
}

// Application =============================

func newApplication(c *clix.Context, producers []streaming.Producer, queueSize int) (*doubleteam.Application, error) {
	app := doubleteam.NewApplication(c, producers, queueSize)

	return app, nil
}

// Producers ===============================

func newKafkaProducer(c *clix.Context) (streaming.Producer, error) {
	brokers := c.StringSlice(FlagKafkaBrokers)
	version := c.String(FlagKafkaVersion)
	retry := c.Int(FlagKafkaRetry)

	return streaming.NewKafkaProducer(brokers, version, retry)
}

func newS3Producer(c *clix.Context) (streaming.Producer, error) {
	endpoint := c.String(FlagS3Endpoint)
	region := c.String(FlagS3Region)
	bucket := c.String(FlagS3Bucket)

	return streaming.NewS3Producer(endpoint, region, bucket)
}

func newS3Consumer(c *clix.Context) (streaming.Consumer, error) {
	endpoint := c.String(FlagS3Endpoint)
	region := c.String(FlagS3Region)
	bucket := c.String(FlagS3Bucket)

	return streaming.NewS3Consumer(endpoint, region, bucket)
}
