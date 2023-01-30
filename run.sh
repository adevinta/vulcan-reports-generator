#!/bin/sh

# Copyright 2021 Adevinta

export PATH_STYLE="${PATH_STYLE:-false}"
export SQS_NUM_PROCESSORS="${SQS_NUM_PROCESSORS:-2}"

envsubst < config.toml > run.toml

if [ ! -z "$PG_CA_B64" ]; then
  mkdir /root/.postgresql
  echo $PG_CA_B64 | base64 -d > /root/.postgresql/root.crt   # for flyway
  echo $PG_CA_B64 | base64 -d > /etc/ssl/certs/pg.crt  # for vulcan-api
fi

# create database if not exists
echo "SELECT 'CREATE DATABASE $PG_NAME' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$PG_NAME')\gexec" | \
PGPASSWORD=$PG_PASSWORD psql -h "$PG_HOST" -p "$PG_PORT" postgres "$PG_USER"

flyway -user="$PG_USER" -password="$PG_PASSWORD" \
  -url="jdbc:postgresql://$PG_HOST:$PG_PORT/$PG_NAME?sslmode=$PG_SSLMODE" \
  -community -baselineOnMigrate=true -locations=filesystem:/app/sql migrate

exec ./vulcan-reports-generator -c run.toml
