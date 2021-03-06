FROM golang:alpine as builder

RUN apk update && \
    apk upgrade && \
    apk add --update curl git

RUN curl -s https://raw.githubusercontent.com/eficode/wait-for/master/wait-for -o /usr/bin/wait-for
RUN chmod +x /usr/bin/wait-for

COPY ./pubsubc /src
WORKDIR /src
RUN go build -o pubsubc


FROM google/cloud-sdk:alpine

COPY --from=builder /usr/bin/wait-for /usr/bin
COPY --from=builder /src/pubsubc /usr/bin
COPY run.sh /run.sh

RUN chmod +x /run.sh

RUN apk add --no-cache --update netcat-openbsd openjdk17-jre-headless && \
    gcloud components install beta pubsub-emulator && \
    gcloud components update

EXPOSE 8681

CMD /run.sh
