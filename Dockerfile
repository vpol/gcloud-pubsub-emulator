FROM --platform=$BUILDPLATFORM golang:bullseye as builder

RUN apt install curl git

COPY ./pubsubc /src
WORKDIR /src
RUN go build -o pubsubc

FROM --platform=$BUILDPLATFORM gcr.io/google.com/cloudsdktool/google-cloud-cli:debian_component_based as runner

RUN curl -s https://raw.githubusercontent.com/eficode/wait-for/master/wait-for -o /usr/local/bin/wait-for
RUN chmod +x /usr/local/bin/wait-for

COPY run.sh /run.sh
RUN chmod +x /run.sh

COPY --from=builder /src/pubsubc /usr/local/bin

RUN apt update && apt install -y netcat-openbsd openjdk-17-jdk-headless

RUN gcloud components install beta pubsub-emulator
RUN gcloud components update

EXPOSE 8681

CMD /run.sh
