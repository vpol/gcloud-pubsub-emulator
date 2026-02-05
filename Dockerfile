FROM --platform=$BUILDPLATFORM golang:alpine AS builder

COPY ./pubsubc /src
WORKDIR /src
RUN go build -o pubsubc

FROM --platform=$BUILDPLATFORM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine AS runner

RUN apk add --no-cache curl openjdk17-jre-headless netcat-openbsd

RUN curl -s https://raw.githubusercontent.com/eficode/wait-for/master/wait-for -o /usr/local/bin/wait-for && \
    chmod +x /usr/local/bin/wait-for

COPY run.sh /run.sh
RUN chmod +x /run.sh

COPY --from=builder /src/pubsubc /usr/local/bin

RUN gcloud components install beta pubsub-emulator && \
    gcloud components update

EXPOSE 8681

CMD /run.sh
