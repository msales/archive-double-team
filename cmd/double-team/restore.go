package main

import (
	"time"

	"github.com/msales/double-team"
	"github.com/msales/double-team/streaming"
	"github.com/msales/pkg/v3/clix"
	"github.com/msales/pkg/v3/log"
	"github.com/msales/pkg/v3/stats"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

func runRestore(c *cli.Context) {
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

	s3Consumer, err := newS3Consumer(ctx)
	if err != nil {
		log.Fatal(ctx, err.Error())
	}

	log.Info(ctx, "Starting restore process")

	messages, errs := s3Consumer.Output(time.Now())
	go func() {
		for err := range errs {
			log.Error(ctx, err.Error())
		}
	}()

	for msgs := range messages {
		for _, msg := range msgs {
			app.Send(msg.Topic, msg.Key, msg.Data)
			stats.Inc(ctx, "consumed", 1, 1.0)
		}

		if err := shouldContinue(app, kafkaProducer); err != nil {
			break
		}
	}

	log.Info(ctx, "Draining queues")

	// Close the application
	if err := app.Close(); err != nil {
		log.Error(ctx, err.Error())
	}

	log.Info(ctx, "Done")
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
