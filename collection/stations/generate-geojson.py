import datetime
import xml.etree.ElementTree as ET
import io
from dataclasses import dataclass
import psycopg
import json

database_connection = psycopg.connect(
    host="localhost",
    port=5432,
    dbname="postgres",
    user="application",
    password="password",
)
database_cursor = database_connection.cursor()

database_cursor.execute(
"""
SELECT * FROM stations;
"""
)
stations = [row for row in database_cursor.fetchall()]


print(stations)


output = {
    "type": "FeatureCollection",
    "features": []
}

for station in stations:
    output["features"].append({
    "type": "Feature",
        "geometry": {
            "type": "Point",
            "coordinates": [station[5], station[4]],
        },
        "properties": {
            "id": station[0],
            "value": 0,
        },
    })


file = open("stations.geojson", "w+")
json.dump(output, file)
file.close()

database_connection.close()