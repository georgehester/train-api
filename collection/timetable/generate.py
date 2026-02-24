import datetime
import xml.etree.ElementTree as ET
import collection.xml
import collection.timetable.parser
from collection.database import Database

timetable_location = "/Volumes/LaCie/train/timetable"
start_date = datetime.date(2026, 1, 1)
end_date = datetime.date(2026, 1, 31)

Database.initialise("/Volumes/LaCie/train.db")

Database.cursor().execute(
"""
SELECT tiploc FROM stations;
"""
)
tiplocs = [row["tiploc"] for row in Database.cursor().fetchall()]

def parse_file(path, date):
    file = open(path, "r")
    data = file.read()
    file.close()

    namespace_dictionary = collection.xml.fetch_namespace_dictionary(data)

    root = ET.fromstring(data)
    collection.timetable.parser.parse_timetable(root, namespace_dictionary, date, tiplocs)

current_date = start_date

while current_date <= end_date:
    path = f"{timetable_location}/{current_date.strftime("%Y-%m-%d")}.xml"

    parse_file(path, current_date)

    current_date += datetime.timedelta(days=1)

Database.close()