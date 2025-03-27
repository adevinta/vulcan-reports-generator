# Copyright 2021 Adevinta

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build ./cmd/vulcan-reports-generator

FROM alpine:3.21

WORKDIR /flyway

RUN apk add --no-cache postgresql-client openjdk17-jre bash gettext libc6-compat

ARG FLYWAY_VERSION=10.10.0

RUN wget -q https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && tar -xzf flyway-commandline-${FLYWAY_VERSION}.tar.gz --strip 1 \
    && rm flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && find ./drivers/ -type f -not -name '*postgres*' -not -name '*jackson*' -delete \
    && chown -R root:root . \
    && ln -s /flyway/flyway /bin/flyway

WORKDIR /app

COPY db/sql /app/sql/

COPY --from=builder /app/vulcan-reports-generator .

# copy generators resources, this is provided by go generate
COPY _build/files/opt/vulcan-reports-generator/generators /app/resources/generators

COPY config.toml .
COPY run.sh .

CMD [ "./run.sh" ]
