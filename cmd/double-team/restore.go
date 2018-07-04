package main

import (
	"log"
	"time"

	"github.com/msales/double-team"
	"github.com/msales/double-team/streaming"
	"github.com/msales/pkg/stats"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

func runRestore(c *cli.Context) {
	ctx, err := newContext(c)
	if err != nil {
		log.Fatal(err.Error())
	}

	go stats.Runtime(ctx.stats)

	kafkaProducer, err := newKafkaProducer(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	s3Producer, err := newS3Producer(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	app, err := newApplication(ctx, []streaming.Producer{kafkaProducer, s3Producer}, c.Int(FlagQueueSize))
	if err != nil {
		log.Fatal(err.Error())
	}

	s3Consumer, err := newS3Consumer(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx.logger.Info("Starting restore process")

	messages, errs := s3Consumer.Output(time.Now())
	go func() {
		for err := range errs {
			ctx.logger.Error(err.Error())
		}
	}()

	for msgs := range messages {
		for _, msg := range msgs {
			app.Send(msg.Topic, msg.Key, msg.Data)
			stats.Inc(ctx, "consumed", 1, 1.0, map[string]string{})
		}

		if err := shouldContinue(app, kafkaProducer); err != nil {
			break
		}
	}

	ctx.logger.Info("Draining queues")

	// Close the application
	if err := app.Close(); err != nil {
		ctx.logger.Error(err.Error())
	}

	ctx.logger.Info("Done")
}

func shouldContinue(app *doubleteam.Application, p streaming.Producer) error {
	if err := app.IsHealthy(); err != nil {
		return err
	}

	if !p.IsHealthy() {
		return errors.New("Producer " + p.Name() + " not healthy")
	}
	return nil
}
