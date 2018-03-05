package main

import (
	"net/url"
	"os"

	"github.com/msales/double-team"
	"github.com/msales/double-team/streaming"
	"github.com/msales/pkg/log"
	"github.com/msales/pkg/stats"
	"gopkg.in/inconshreveable/log15.v2"
)

// Application =============================

func newApplication(c *Context, producers []streaming.Producer, queueSize int) (*doubleteam.Application, error) {
	app := doubleteam.NewApplication(c, producers, queueSize)

	return app, nil
}

// Producers ===============================

func newKafkaProducer(c *Context) (streaming.Producer, error) {
	brokers := c.StringSlice(FlagKafkaBrokers)
	retry := c.Int(FlagKafkaRetry)

	return streaming.NewKafkaProducer(brokers, retry)
}

func newS3Producer(c *Context) (streaming.Producer, error) {
	endpoint := c.String(FlagS3Endpoint)
	region := c.String(FlagS3Region)
	bucket := c.String(FlagS3Bucket)

	return streaming.NewS3Producer(endpoint, region, bucket)
}

func newS3Consumer(c *Context) (streaming.Consumer, error) {
	endpoint := c.String(FlagS3Endpoint)
	region := c.String(FlagS3Region)
	bucket := c.String(FlagS3Bucket)

	return streaming.NewS3Consumer(endpoint, region, bucket)
}

// Logger ==================================

func newLogger(c *Context) (log15.Logger, error) {
	lvl := c.String(FlagLogLevel)
	v, err := log15.LvlFromString(lvl)
	if err != nil {
		return nil, err
	}

	h := log15.LvlFilterHandler(v, log15.StreamHandler(os.Stderr, log15.LogfmtFormat()))
	if lvl == "debug" {
		h = log15.CallerFileHandler(h)
	}

	l := log15.New()
	l.SetHandler(log15.LazyHandler(h))

	return l, nil
}

// Stats ===================================

func newStats(c *Context) (stats.Stats, error) {
	dsn := c.String(FlagStats)
	if dsn == "" {
		return stats.Null, nil
	}

	uri, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	switch uri.Scheme {
	case "statsd":
		return newStatsdStats(uri.Host)

	case "l2met":
		return newL2metStats(c.logger), nil

	default:
		return stats.Null, nil
	}
}

func newStatsdStats(addr string) (stats.Stats, error) {
	return stats.NewStatsd(addr, "double-team")
}

func newL2metStats(log log.Logger) stats.Stats {
	return stats.NewL2met(log, "double-team")
}
