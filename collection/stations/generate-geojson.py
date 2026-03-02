from collection.database import Database
import json

Database.initialise("localhost", 5432, "train", "application", "password")

cursor = Database.execute("SELECT * FROM stations;")
stations = [row for row in cursor.fetchall()]

cursor = Database.execute(
"""
SELECT stations.tiploc, COUNT(stops.id) as count FROM stations
LEFT JOIN stops ON stops.station_tiploc = stations.tiploc
GROUP BY stations.tiploc;
""")
service_count = {row["tiploc"]: row["count"] for row in cursor.fetchall()}

output = {
    "type": "FeatureCollection",
    "features": []
}

for station in stations:
    output["features"].append({
    "type": "Feature",
        "geometry": {
            "type": "Point",
            "coordinates": [station["longitude"], station["latitude"]],
        },
        "properties": {
            "id": station["tiploc"],
            "value": service_count[station["tiploc"]],
        },
    })

file = open("stations.geojson", "w+")
json.dump(output, file)
file.close()

Database.close()