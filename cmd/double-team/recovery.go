package main

import (
	"log"
	"time"

	"github.com/msales/double-team"
	"github.com/msales/double-team/producer"
	"github.com/msales/pkg/stats"
	"gopkg.in/urfave/cli.v1"
	"github.com/pkg/errors"
)

func runRecovery(c *cli.Context) {
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

	app, err := newApplication(ctx, []producer.Producer{kafkaProducer, s3Producer}, c.Int(FlagQueueSize))
	if err != nil {
		log.Fatal(err.Error())
	}

	s3Consumer, err := newS3Consumer(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx.logger.Info("Starting recovery process")

	for msgs := range s3Consumer.Output(time.Now()) {
		for _, msg := range msgs {
			app.Send(msg.Topic, msg.Data)
		}

		if err := shouldContinue(app, kafkaProducer); err != nil {
			log.Fatal(err)
		}
	}

	ctx.logger.Info("Draining queues")

	// Close the application
	if err := app.Close(); err != nil {
		ctx.logger.Error(err.Error())
	}

	ctx.logger.Info("Done")
}

func shouldContinue(app *doubleteam.Application, p producer.Producer) error {
	if err := app.IsHealthy(); err != nil {
		return err
	}

	if !p.IsHealthy() {
		return errors.New("Producer " + p.Name() + " not healthy")
	}
	return nil
}