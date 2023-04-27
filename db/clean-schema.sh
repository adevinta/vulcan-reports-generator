#!/usr/bin/env bash

# Copyright 2021 Adevinta

docker exec -i -e PGUSER=vulcan_reportgen -e PGPASSWORD=vulcan_reportgen vulcan-reportgen-db \
    psql -h localhost reportsgenerator -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"
