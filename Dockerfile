FROM --platform=$BUILDPLATFORM golang:bullseye as builder

RUN apt install curl git

RUN curl -s https://raw.githubusercontent.com/eficode/wait-for/master/wait-for -o /usr/bin/wait-for
RUN chmod +x /usr/bin/wait-for

COPY ./pubsubc /src
WORKDIR /src
RUN go build -o pubsubc


FROM --platform=$BUILDPLATFORM google/cloud-sdk:debian_component_based

COPY --from=builder /usr/bin/wait-for /usr/local/bin
COPY --from=builder /src/pubsubc /usr/local/bin
COPY run.sh /run.sh

RUN chmod +x /run.sh

RUN apt update && apt install -y netcat-openbsd openjdk-11-jdk-headless

RUN gcloud components install beta pubsub-emulator
RUN gcloud components update

EXPOSE 8681

CMD /run.sh
