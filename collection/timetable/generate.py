import datetime
import xml.etree.ElementTree as ET
import collection.xml
import collection.timetable.parser
from collection.database import Database
import concurrent.futures

timetable_location = "/Volumes/LaCie/train/timetable"
start_date = datetime.date(2026, 1, 1)
end_date = datetime.date(2026, 1, 31)

Database.initialise("localhost", 5432, "train", "application", "password")

cursor = Database.execute("SELECT tiploc FROM stations;")
tiplocs = [row["tiploc"] for row in cursor.fetchall()]

Database.execute("DELETE FROM stops;")
Database.commit()

def parse_file(path, date):
    file = open(path, "r")
    data = file.read()
    file.close()

    namespace_dictionary = collection.xml.fetch_namespace_dictionary(data)
    root = ET.fromstring(data)

    collection.timetable.parser.parse_timetable(root, namespace_dictionary, date, tiplocs)

dates = []
current_date = start_date

while current_date <= end_date:
    dates.append(current_date)
    current_date += datetime.timedelta(days=1)

with concurrent.futures.ThreadPoolExecutor(max_workers=8) as executor:
    futures = {
        executor.submit(
            parse_file,
            f"{timetable_location}/{date.strftime("%Y-%m-%d")}.xml",
            date
        ): date
        for date in dates
    }

    for future in concurrent.futures.as_completed(futures):
        date = futures[future]

        try:
            future.result()
            print(f"[ Success ][ Completed Generation For {date} ]")
        except:
            raise RuntimeError(f"[ Error ][ Failed Generation For {date} ]")

Database.close()