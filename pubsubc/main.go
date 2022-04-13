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

func create(ctx context.Context, project string, subscription string, topic string) error {

	logger := log.With().Str("project", project).Str("subscription", subscription).Str("topic", topic).Logger()

	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("unable to create client to project %q: %s", project, err)
	}

	defer func() {
		client.Close()
		<-time.After(3 * time.Second)
	}()

	logger.Trace().Msg("client connected")

	t, err := client.CreateTopic(ctx, topic)
	if err != nil {
		if !strings.Contains(err.Error(), "AlreadyExists") {
			return fmt.Errorf("unable to create topic %q for project %q: %s", topic, project, err)
		}

		t = client.Topic(topic)
	}

	logger.Trace().Msg("topic created")

	_, err = client.CreateSubscription(ctx, subscription, pubsub.SubscriptionConfig{Topic: t})
	if err != nil {
		return fmt.Errorf("unable to create subscription %q on topic %q for project %q: %s", subscription, topic, project, err)
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

	for _, v := range config.Subscriptions {
		err := create(context.TODO(), v.Project, v.Subscription, v.Topic)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to process subscription")

		}
	}

}
