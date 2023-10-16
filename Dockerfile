# Copyright 2021 Adevinta

FROM golang:1.19-alpine3.18 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd cmd/vulcan-reports-generator/ && GOOS=linux GOARCH=amd64 go build . && cd -

FROM alpine:3.18.3

WORKDIR /flyway

# add psql client to create DB from run script
RUN apk add postgresql-client

# add flyway
RUN apk add --no-cache --update openjdk8-jre-base bash gettext libc6-compat

ARG FLYWAY_VERSION=9.19.3

RUN wget https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && tar -xzf flyway-commandline-${FLYWAY_VERSION}.tar.gz --strip 1 \
    && rm flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && find ./drivers/ -type f -not -name 'postgres*' -delete \
    && chown -R root:root . \
    && ln -s /flyway/flyway /bin/flyway

ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"

ENV BUILD_RFC3339 "$BUILD_RFC3339"
ENV COMMIT "$COMMIT"

WORKDIR /app

COPY db/sql /app/sql/

COPY --from=builder /app/cmd/vulcan-reports-generator/vulcan-reports-generator .

# copy generators resources, this is provided by go generate
COPY _build/files/opt/vulcan-reports-generator/generators /app/resources/generators

COPY config.toml .
COPY run.sh .

CMD [ "./run.sh" ]
