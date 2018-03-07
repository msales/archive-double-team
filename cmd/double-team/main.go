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

	FlagQueueSize = "queue"

	FlagKafkaBrokers = "kafka.brokers"
	FlagKafkaRetry   = "kafka.retry"

	FlagS3Endpoint = "s3.endpoint"
	FlagS3Region   = "s3.region"
	FlagS3Bucket   = "s3.bucket"
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
		Usage: "Run the HTTP server",
		Flags: append([]cli.Flag{
			cli.IntFlag{
				Name:   FlagQueueSize,
				Value:  1000,
				Usage:  "The queue size of the message buffers.",
				EnvVar: "DOUBLE_TEAM_QUEUE",
			},
			cli.StringSliceFlag{
				Name:   FlagKafkaBrokers,
				Usage:  "The kafka seed brokers.",
				EnvVar: "DOUBLE_TEAM_KAFKA_BROKERS",
			},
			cli.IntFlag{
				Name:   FlagKafkaRetry,
				Value:  5,
				Usage:  "The number of times to retry producing a message.",
				EnvVar: "DOUBLE_TEAM_KAFKA_RETRY",
			},
			cli.StringFlag{
				Name:   FlagS3Endpoint,
				Value:  "",
				Usage:  "The s3 endpoint. Only set for testing.",
				EnvVar: "DOUBLE_TEAM_S3_ENDPOINT",
			},
			cli.StringFlag{
				Name:   FlagS3Region,
				Usage:  "The s3 bucket region.",
				EnvVar: "DOUBLE_TEAM_S3_REGION",
			},
			cli.StringFlag{
				Name:   FlagS3Bucket,
				Usage:  "The s3 bucket.",
				EnvVar: "DOUBLE_TEAM_S3_BUCKET",
			},
			cli.StringFlag{
				Name:   FlagPort,
				Value:  "80",
				Usage:  "The port to run the server on.",
				EnvVar: "DOUBLE_TEAM_PORT",
			},
		}, commonFlags...),
		Action: runServer,
	},
	{
		Name:  "restore",
		Usage: "Restore from S3",
		Flags: append([]cli.Flag{
			cli.IntFlag{
				Name:   FlagQueueSize,
				Value:  1000,
				Usage:  "The queue size of the message buffers.",
				EnvVar: "DOUBLE_TEAM_QUEUE",
			},
			cli.StringSliceFlag{
				Name:   FlagKafkaBrokers,
				Usage:  "The kafka seed brokers.",
				EnvVar: "DOUBLE_TEAM_KAFKA_BROKERS",
			},
			cli.IntFlag{
				Name:   FlagKafkaRetry,
				Value:  5,
				Usage:  "The number of times to retry producing a message.",
				EnvVar: "DOUBLE_TEAM_KAFKA_RETRY",
			},
			cli.StringFlag{
				Name:   FlagS3Endpoint,
				Value:  "",
				Usage:  "The s3 endpoint. Only set for testing.",
				EnvVar: "DOUBLE_TEAM_S3_ENDPOINT",
			},
			cli.StringFlag{
				Name:   FlagS3Region,
				Usage:  "The s3 bucket region.",
				EnvVar: "DOUBLE_TEAM_S3_REGION",
			},
			cli.StringFlag{
				Name:   FlagS3Bucket,
				Usage:  "The s3 bucket.",
				EnvVar: "DOUBLE_TEAM_S3_BUCKET",
			},
		}, commonFlags...),
		Action: runRestore,
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "double-team"
	app.Usage = "An HTTP Kafka producer"
	app.Version = Version
	app.Commands = commands

	app.Run(os.Args)
}
