# Setting-up Golang custom Mapserver

## Download OSM (Open Source Map) from below website

```bash
curl https://download.geofabrik.de/asia/india-latest.osm.pbf -o ./india-latest.osm.pbf
```

## Install Postgres and Post-gis

[check official page](https://www.postgis.net/workshops/postgis-intro/installation.html#)

## Migrate pbf map data into postgres

1. Create Database

```Sql
CREATE DATABASE [DATABASE_NAME];
```

2. Enable Post-Gis

```sql
\c [DATABASE_NAME] -- use the database

CREATE EXTENSION postgis; -- enable the extension

SELECT postgis_full_version(); -- VERIFY installation of post gis
```

3. Migrate Map data into postgresDB

   [checkout the official page](https://www.cybertec-postgresql.com/en/open-street-map-to-postgis-the-basics/)

4. Build the Backend go code and run

# Backend

## Install depedencies

```bash
go mod tidy
```

## Clone sample env

```bash
cp ./server/example.env ./server/.env
```

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB Container

```bash
make docker-down
```

DB Integrations Test:

```bash

```

Live reload the application:

```bash
make watch
```

Run the test suite:

```bash
make test
```

Clean up binary from the last build:

```bash
make clean
```

# Client

## install node packages

```bash
 cd client && yarn
```

## Run client app

```bash
yarn run dev
```
