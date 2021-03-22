#!/bin/sh

# Copyright 2021 Adevinta

source postgres-stop.sh
source postgres-start.sh
sleep 2
source flyway-migrate.sh
