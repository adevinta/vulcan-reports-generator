#!/usr/bin/env bash

# Copyright 2021 Adevinta

docker run --net=host -v "$PWD":/flyway/sql flyway/flyway:"${FLYWAY_VERSION:-10}-alpine" \
    -user=vulcan_reportgen -password=vulcan_reportgen -url=jdbc:postgresql://localhost:5438/vulcan_reportgen -baselineOnMigrate=true -cleanDisabled=false clean
