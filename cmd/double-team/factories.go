package main

import (
	"net/url"
	"os"

	"github.com/msales/double-team"
	"github.com/msales/pkg/log"
	"github.com/msales/pkg/stats"
	"gopkg.in/inconshreveable/log15.v2"
)

// Application =============================

func newApplication(c *Context) (*double_team.Application, error) {
	app := double_team.NewApplication()

	return app, nil
}

// Logger ==================================

func newLogger(c *Context) (log15.Logger, error) {
	lvl := c.String(FlagLogLevel)
	v, err := log15.LvlFromString(lvl)
	if err != nil {
		return nil, err
	}

	h := log15.LvlFilterHandler(v, log15.StreamHandler(os.Stderr, log15.TerminalFormat()))
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
	return stats.NewStatsd(addr, "ren")
}

func newL2metStats(log log.Logger) stats.Stats {
	return stats.NewL2met(log, "ren")
}
