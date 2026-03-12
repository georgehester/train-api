# Table of Contents

- [Table of Contents](#table-of-contents)
- [Contributing](#contributing)
- [Setup](#setup)
  - [Generation Of Signing Keys](#generation-of-signing-keys)
  - [Generation Of OpenAPI Documentation](#generation-of-openapi-documentation)
  - [Generation Of MBTiles](#generation-of-mbtiles)
  - [Execution](#execution)

# Contributing

Refer to the [Contributing Guide](CONTRIBUTING.md) for detailed information about commit messages.


# Setup

## Generation Of Signing Keys

To generate the keys we use to sign JWT (JSON Web Tokens) use the `generate-keys.sh` script which will create a new Ed25519 key pair and Base64 encode them for use in our Docker Compose.

```sh
./generate-keys.sh
```

## Generation Of OpenAPI Documentation

To generate the OpenAPI JSON documentation file simply run the following command.

```sh
./swagger.sh
```

## Generation Of MBTiles

Our tile server requires MBTiles files relating to the railway lines and the outline of Great Britain as defined in the `data/configuration.json` file.

To start with we can gather the railway data. To do this we will use the [OpenRailwayMap Vector Map](https://github.com/hiddewie/OpenRailwayMap-vector) repository. Start by cloning the repository.

```sh
git clone git@github.com:hiddewie/OpenRailwayMap-vector.git
```

Once cloned we need to populate the `data` folder with an `.osm.pbf` file. We will use [Geofabrik](https://download.geofabrik.de/europe/united-kingdom.html) for this, downloading the United Kingdom region and renaming it to `data.osm.pbf`.

Once complete we can build the OpenRailwayMap service and deploy it in Docker.

```sh
docker compose up --build --watch db import martin proxy api
```

> [!NOTE]
> You may need to run `git config core.precomposeunicode true` to fix unicode errors.

We now want to generate MBTiles from the database produced by OpenRailwayMap. To do this we can make use of the Martin tile server. We want to extract both the low and high information versions of the map to show at different zoom levels so run the following two commands.

```sh
martin-cp \
  --output-file railway-line-low.mbtiles \
  --mbtiles-type normalized \
  --bbox="-8.65,49.86,1.77,60.86" \
  --min-zoom 0 \
  --max-zoom 14 \
  "postgresql://postgres:postgres@localhost:5432/gis" \
  --source "railway_line_low"
```

```sh
martin-cp \
  --output-file railway-line-high.mbtiles \
  --mbtiles-type normalized \
  --bbox="-8.65,49.86,1.77,60.86" \
  --min-zoom 0 \
  --max-zoom 14 \
  "postgresql://postgres:postgres@localhost:5432/gis" \
  --source "railway_line_high"
```

Now we have the `railway-line-high.mbtiles` and `railway-line-low.mbtiles` files we can place them in our local `data` folder for use by our tile server.

Next we want to focus on the `great-britain.mbtiles` file. The information required for this generation is preconfigured in our `data/planetiler.yaml` file. To generate the MBTiles we can just run the Planetiler Docker with the following command.

```sh
./planetiler.sh
```

## Execution

The application and its components are entirely hosted inside of Docker and are defined using the Docker Compose configuration file. To run the application use the following command.

```sh
docker compose up -d
```

To stop the applications you can run.

```sh
docker compose down
```