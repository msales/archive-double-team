package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/msales/pkg/v3/clix"
	"gopkg.in/urfave/cli.v1"
)

// Flag constants declared for CLI use.
const (
	FlagQueueSize = "queue"

	FlagKafkaBrokers = "kafka.brokers"
	FlagKafkaVersion = "kafka.version"
	FlagKafkaRetry   = "kafka.retry"

	FlagS3Endpoint = "s3.endpoint"
	FlagS3Region   = "s3.region"
	FlagS3Bucket   = "s3.bucket"
)

var flags = clix.Flags{
	cli.IntFlag{
		Name:   FlagQueueSize,
		Value:  1000,
		Usage:  "The queue size of the message buffers.",
		EnvVar: "DOUBLE_TEAM_QUEUE",
	},
}

var s3Flags = clix.Flags{
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
}

var kafkaFlags = clix.Flags{
	cli.StringSliceFlag{
		Name:   FlagKafkaBrokers,
		Usage:  "The kafka seed brokers.",
		EnvVar: "DOUBLE_TEAM_KAFKA_BROKERS",
	},
	cli.StringFlag{
		Name:   FlagKafkaVersion,
		Usage:  "Kafka version.",
		EnvVar: "DOUBLE_TEAM_KAFKA_VERSION",
	},
	cli.IntFlag{
		Name:   FlagKafkaRetry,
		Value:  5,
		Usage:  "The number of times to retry producing a message.",
		EnvVar: "DOUBLE_TEAM_KAFKA_RETRY",
	},
}

var commands = []cli.Command{
	{
		Name:  "server",
		Usage: "Run the HTTP server",
		Flags: clix.Flags.Merge(
			clix.CommonFlags,
			clix.ServerFlags,
			s3Flags,
			kafkaFlags,
			flags,
		),
		Action: runServer,
	},
	{
		Name:  "restore",
		Usage: "Restore from S3",
		Flags: clix.Flags.Merge(
			clix.CommonFlags,
			s3Flags,
			kafkaFlags,
			flags,
		),
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
