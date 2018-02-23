package main

import (
	"os"

	"gopkg.in/urfave/cli.v1"
)

import (
	_ "github.com/joho/godotenv/autoload"
)

// Flag constants declared for CLI use.
const (
	FlagPort     = "port"
	FlagLogLevel = "log.level"

	FlagStats = "stats"
)

var commonFlags = []cli.Flag{
	cli.StringFlag{
		Name:   FlagLogLevel,
		Value:  "info",
		Usage:  "Specify the log level. You can use this to enable debug logs by specifying `debug`.",
		EnvVar: "DOUBLE_TEAM_LOG_LEVEL",
	},
	cli.StringFlag{
		Name:   FlagStats,
		Value:  "",
		Usage:  "The stats backend to use. (e.g. statsd://localhost:8125)",
		EnvVar: "DOUBLE_TEAM_STATS",
	},
}

var commands = []cli.Command{
	{
		Name:  "server",
		Usage: "Run the ren HTTP server",
		Flags: append([]cli.Flag{
			cli.StringFlag{
				Name:   FlagPort,
				Value:  "80",
				Usage:  "The port to run the server on.",
				EnvVar: "DOUBLE_TEAM_PORT",
			},

		}, commonFlags...),
		Action: runServer,
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "ren"
	app.Version = Version
	app.Commands = commands

	app.Run(os.Args)
}
