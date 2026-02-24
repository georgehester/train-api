import psycopg
import json
from collection.database import Database

station_location = "/Volumes/LaCie/train/station"

Database.initialise("/Volumes/LaCie/train.db")

with open(f"{station_location}/stations.json", "r") as file:
    stations_data = json.load(file)

with open(f"{station_location}/corpus.json", "r") as file:
    corpus_data = json.load(file)

for station in stations_data["stations"]:
    corpus_match = next(
        (
            item
            for item in corpus_data["TIPLOCDATA"]
            if item["3ALPHA"] == station["crsCode"]
        ),
        None,
    )

    if corpus_match == None:
        raise RuntimeError(f"[ Error ][ Could not find TIPLOC for CRS {station["crsCode"]} ]")

    Database.cursor().execute(
        """
        INSERT INTO
            stations (name, crs, nlc, tiploc, latitude, longitude)
        VALUES
            (?, ?, ?, ?, ?, ?);
        """,
        (
            station["name"],
            station["crsCode"],
            station["nationalLocationCode"],
            corpus_match["TIPLOC"],
            station["location"]["latitude"],
            station["location"]["longitude"],
        ),
    )
    Database.connection().commit()

Database.close()