# Copyright 2021 Adevinta

FROM golang:1.19.3-alpine3.15 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd cmd/vulcan-reports-generator/ && GOOS=linux GOARCH=amd64 go build . && cd -

FROM golang:1.19.3-alpine3.15 as migrations

RUN GOBIN=/ go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:3.17.3

RUN apk add --no-cache --update bash gettext libc6-compat postgresql-client

ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"

ENV BUILD_RFC3339 "$BUILD_RFC3339"
ENV COMMIT "$COMMIT"

WORKDIR /app

COPY --from=migrations /migrate /bin
COPY --from=builder /app/cmd/vulcan-reports-generator/vulcan-reports-generator .
COPY db/migrations /app/migrations

# copy generators resources, this is provided by go generate
COPY _build/files/opt/vulcan-reports-generator/generators /app/resources/generators

COPY config.toml .
COPY run.sh .

CMD [ "./run.sh" ]
