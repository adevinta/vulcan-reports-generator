#!/usr/bin/env bash

# Copyright 2021 Adevinta

docker run --net=host -v "$PWD":/flyway/sql flyway/flyway:"${FLYWAY_VERSION:-8}-alpine" \
    -user=vulcan_reportgen -password=vulcan_reportgen -url=jdbc:postgresql://localhost:5438/vulcan_reportgen -baselineOnMigrate=true clean
