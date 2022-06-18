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

func create(ctx context.Context, log zerolog.Logger, client *pubsub.Client, subscription string, topic string) error {

	logger := log.With().Str("subscription", subscription).Str("topic", topic).Logger()

	logger.Trace().Msg("client connected")

	t, err := client.CreateTopic(ctx, topic)
	if err != nil {
		if !strings.Contains(err.Error(), "AlreadyExists") {
			return fmt.Errorf("unable to create topic %q: %w", topic, err)
		}

		t = client.Topic(topic)
	}

	logger.Trace().Msg("topic created")

	_, err = client.CreateSubscription(ctx, subscription, pubsub.SubscriptionConfig{Topic: t})
	if err != nil {
		return fmt.Errorf("unable to create subscription %q on topic %q: %w", subscription, topic, err)
	}

	logger.Trace().Msg("subscription created")

	return nil
}

type Subscription struct {
	Project      string `toml:"project"`
	Subscription string `toml:"subscription"`
	Topic        string `toml:"topic"`
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

		err := create(context.TODO(), logger, c, v.Subscription, v.Topic)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to process subscription")
		}

	}
}
