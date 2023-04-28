#!/bin/sh

# Copyright 2021 Adevinta

# build-model.sh generates the ORM model data
# based on the defined DB schema.

# get boilersql
go get -u -t github.com/volatiletech/sqlboiler@v3.6.1
go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql@v3.6.1

# start db
source postgres-start.sh
sleep 2
source migrate.sh

# generate model in tmp gen folder
# copy it to pkg/model and clean
mkdir sqlboiler/gen
sqlboiler -c sqlboiler/sqlboiler.toml --wipe psql
mv -f sqlboiler/gen/* ../pkg/storage
rm -rf sqlboiler/gen
