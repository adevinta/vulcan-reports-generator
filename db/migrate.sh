#!/usr/bin/env bash

# Copyright 2021 Adevinta

docker run -v $PWD/pkg/migrations:/migrations --network host migrate/migrate \
    -path=/migrations -database='postgres://vulcan_reportgen:vulcan_reportgen@localhost:5438/vulcan_reportgen?sslmode=disable' -verbose up
