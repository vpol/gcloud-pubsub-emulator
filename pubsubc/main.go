package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

func create(ctx context.Context, log zerolog.Logger, client *pubsub.Client, sub Subscription) error {

	logger := log.With().Str("subscription", sub.Subscription).Str("topic", sub.Topic).Logger()

	logger.Trace().Msg("client connected")

	t, err := client.CreateTopic(ctx, sub.Topic)
	if err != nil {
		if !strings.Contains(err.Error(), "AlreadyExists") {
			return fmt.Errorf("unable to create topic %q: %w", sub.Topic, err)
		}

		t = client.Topic(sub.Topic)
	}

	logger.Trace().Msg("topic created")

	cfg := pubsub.SubscriptionConfig{Topic: t}

	ackDeadline := 60
	if sub.AckDeadlineSeconds > 0 {
		ackDeadline = sub.AckDeadlineSeconds
	}
	cfg.AckDeadline = time.Duration(ackDeadline) * time.Second

	if cfg.RetryPolicy == nil {
		cfg.RetryPolicy = &pubsub.RetryPolicy{}
	}

	minBackoff := 10
	if sub.MinBackoff > 0 && sub.MinBackoff < 600 {
		minBackoff = sub.MinBackoff
	}
	cfg.RetryPolicy.MinimumBackoff = time.Duration(minBackoff) * time.Second

	maxbackoff := 10
	if sub.MaxBackoff > 0 && sub.MaxBackoff < 600 {
		maxbackoff = sub.MaxBackoff
	}

	cfg.RetryPolicy.MaximumBackoff = time.Duration(maxbackoff) * time.Second

	_, err = client.CreateSubscription(ctx, sub.Subscription, cfg)
	if err != nil {
		return fmt.Errorf("unable to create subscription %q on topic %q: %w", sub.Subscription, sub.Topic, err)
	}

	logger.Trace().Msg("subscription created")

	return nil
}

type Subscription struct {
	Project            string `toml:"project"`
	Subscription       string `toml:"subscription"`
	Topic              string `toml:"topic"`
	AckDeadlineSeconds int    `toml:"ackdeadline"`
	MinBackoff         int    `toml:"minbackoff"`
	MaxBackoff         int    `toml:"maxbackoff"`
}

type ConfigFile struct {
	Subscriptions []Subscription `toml:"subscription"`
}

type Config struct {
	ConfigFile string `env:"CONFIG_FILE,required"`
	LogLevel   string `env:"LOGLEVEL" envDefault:"debug"`
}

func main() {

	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config")
		return
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse loglevel")
		return
	}

	zerolog.SetGlobalLevel(logLevel)

	var config ConfigFile
	_, err = toml.DecodeFile(cfg.ConfigFile, &config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to decode file")
	}

	ctx := context.TODO()

	var connectors = make(map[string]*pubsub.Client)
	defer func() {
		// close them all
		for _, c := range connectors {
			_ = c.Close()
		}
		time.Sleep(3 * time.Second)
	}()

	for _, v := range config.Subscriptions {

		c, ok := connectors[v.Project]
		if !ok {
			// if no connector for project - connect
			c, err = pubsub.NewClient(ctx, v.Project)
			if err != nil {
				log.Panic().Err(err).Str("project", v.Project).Msg("unable to create client")
			}
			connectors[v.Project] = c
		}

		logger := log.With().Str("project", v.Project).Logger()

		err := create(context.TODO(), logger, c, v)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to process subscription")
		}

	}
}
