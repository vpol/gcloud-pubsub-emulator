# gcloud-pubsub-emulator

How to use:

Write config, ie. ./gpubsub/config.toml

```toml
[[subscription]]
project = "sample-project"
subscription = "sample-sub"
topic = "sample-topic"
```


Add it to your docker compose

```
  gpubsub:
    restart: on-failure
    image: vpol/gcloud-pubsub-emulator:latest
    environment:
      - CONFIG_FILE=/conf/config.toml
      - LOGLEVEL=trace
    volumes:
      - ./gpubsub:/conf
    healthcheck:
      test: [ "CMD", "nc", "-z", "gpubsub", "8682" ]
      start_period: 5s
      interval: 2s
      timeout: 1s
      retries: 20
    expose:
      - "8682:8682"
    ports:
      - "8681:8681"
```

Use it as dependency for you magic service

```
  debezium:
    restart: on-failure
    image: debezium/server:2.0.0.Beta1
    volumes:
      - ./debezium/conf:/debezium/conf
      - ./debezium/data:/debezium/data
    depends_on:
      gpubsub:
        condition: service_healthy
```

Run

```
➜  docker-compose run gpubsub
Executing: /google-cloud-sdk/platform/pubsub-emulator/bin/cloud-pubsub-emulator --host=0.0.0.0 --port=8681
[pubsub] This is the Google Pub/Sub fake.
[pubsub] Implementation may be incomplete or differ from the real system.
[pubsub] Jun 10, 2022 2:39:47 AM com.google.cloud.pubsub.testing.v1.Main main
[pubsub] INFO: IAM integration is disabled. IAM policy methods and ACL checks are not supported
[pubsub] SLF4J: Failed to load class "org.slf4j.impl.StaticLoggerBinder".
[pubsub] SLF4J: Defaulting to no-operation (NOP) logger implementation
[pubsub] SLF4J: See http://www.slf4j.org/codes.html#StaticLoggerBinder for further details.
[pubsub] Jun 10, 2022 2:39:52 AM com.google.cloud.pubsub.testing.v1.Main main
[pubsub] INFO: Server started, listening on 8681
{"level":"trace","project":"sample-project","subscription":"sample-sub","topic":"sample-topic","time":"2022-06-10T02:39:53Z","message":"client connected"}
[pubsub] Jun 10, 2022 2:39:53 AM io.gapi.emulators.netty.HttpVersionRoutingHandler channelRead
[pubsub] INFO: Detected HTTP/2 connection.
{"level":"trace","project":"sample-project","subscription":"sample-sub","topic":"sample-topic","time":"2022-06-10T02:39:54Z","message":"topic created"}
{"level":"trace","project":"sample-project","subscription":"sample-sub","topic":"sample-topic","time":"2022-06-10T02:39:54Z","message":"subscription created"}
```

voila!
