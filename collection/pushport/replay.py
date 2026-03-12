import datetime
import xml.etree.ElementTree as ET
import collection.xml
from collection.database import Database
from collection.parser import parse_date, parse_boolean
import collection.pushport.classes

darwin_location =  "/Volumes/LaCie/train/darwin"

start_date = datetime.date(2026, 1, 1)
end_date = datetime.date(2026, 2, 1)

Database.initialise("localhost", 5432, "train", "application", "password")

def chunk_file(path, delimiter="<?xml version=\"1.0\" encoding=\"utf-8\"?>"):
    with open(path, "r", encoding="utf-8") as file:
        buffer = []

        for line in file:
            if delimiter in line and buffer:
                before, after = line.split(delimiter, 1)
                if before:
                    buffer.append(before)

                yield "".join(buffer)

                buffer = [delimiter + after]
            else:
                buffer.append(line)

        if buffer:
            yield "".join(buffer)

def parse_arrival(record: ET.Element, namespace_dictionary: dict[str, str], location: collection.pushport.classes.Location, cursor):
    time = collection.pushport.classes.Time(
        location=location,
        estimatedTime=record.get("et"),
        workingEstimatedTime=record.get("wet"),
        actualTime=record.get("at"),
        actualRemoved=parse_boolean(record.get("atRemoved"), False),
        actualClass=record.get("atClass"),
        estimatedMinimum=record.get("etmin"),
        estimatedUnknown=parse_boolean(record.get("etUnknown"), False),
        isDelayed=parse_boolean(record.get("delayed"), False),
        source=record.get("src"),
        sourceInstance=record.get("srcInst")
    )

    if time.actualRemoved == True:
        cursor.execute(
        """
        UPDATE
            stops
        SET
            actual_arrival = NULL
        WHERE
            station_tiploc = %s AND journey_id = %s;
        """,
        (time.location.locationTIPLOC, time.location.trainStatus.rid))

    if time.actualTime != None:
        cursor.execute(
        """
        UPDATE
            stops
        SET
            actual_arrival = %s
        WHERE
            station_tiploc = %s AND journey_id = %s;
        """,
        (time.actualTime, time.location.locationTIPLOC, time.location.trainStatus.rid))

def parse_departure(record: ET.Element, namespace_dictionary: dict[str, str], location: collection.pushport.classes.Location, cursor):
    time = collection.pushport.classes.Time(
        location=location,
        estimatedTime=record.get("et"),
        workingEstimatedTime=record.get("wet"),
        actualTime=record.get("at"),
        actualRemoved=parse_boolean(record.get("atRemoved"), False),
        actualClass=record.get("atClass"),
        estimatedMinimum=record.get("etmin"),
        estimatedUnknown=parse_boolean(record.get("etUnknown"), False),
        isDelayed=parse_boolean(record.get("delayed"), False),
        source=record.get("src"),
        sourceInstance=record.get("srcInst")
    )

    if time.actualRemoved == True:
        cursor.execute(
        """
        UPDATE
            stops
        SET
            actual_departure = NULL
        WHERE
            station_tiploc = %s AND journey_id = %s;
        """,
        (time.location.locationTIPLOC, time.location.trainStatus.rid))

    if time.actualTime != None:
        cursor.execute(
        """
        UPDATE
            stops
        SET
            actual_departure = %s
        WHERE
            station_tiploc = %s AND journey_id = %s;
        """,
        (time.actualTime, time.location.locationTIPLOC, time.location.trainStatus.rid))

def parse_passing(record: ET.Element, namespace_dictionary: dict[str, str], location: collection.pushport.classes.Location, cursor):
    time = collection.pushport.classes.Time(
        location=location,
        estimatedTime=record.get("et"),
        workingEstimatedTime=record.get("wet"),
        actualTime=record.get("at"),
        actualRemoved=parse_boolean(record.get("atRemoved"), False),
        actualClass=record.get("atClass"),
        estimatedMinimum=record.get("etmin"),
        estimatedUnknown=parse_boolean(record.get("etUnknown"), False),
        isDelayed=parse_boolean(record.get("delayed"), False),
        source=record.get("src"),
        sourceInstance=record.get("srcInst")
    )

    if time.actualRemoved == True:
        cursor.execute(
        """
        UPDATE
            stops
        SET
            actual_passing = NULL
        WHERE
            station_tiploc = %s AND journey_id = %s;
        """,
        (time.location.locationTIPLOC, time.location.trainStatus.rid))

    if time.actualTime != None:
        cursor.execute(
        """
        UPDATE
            stops
        SET
            actual_passing = %s
        WHERE
            station_tiploc = %s AND journey_id = %s;
        """,
        (time.actualTime, time.location.locationTIPLOC, time.location.trainStatus.rid))

def parse_location(record: ET.Element, namespace_dictionary: dict[str, str], train_status: collection.pushport.classes.TrainStatus, cursor):
    location = collection.pushport.classes.Location(
        trainStatus = train_status,
        locationTIPLOC = record.get("tpl")
    )

    arrival_records = record.findall("namespace:arr", {"namespace": namespace_dictionary.get("fc")})
    for arrival in arrival_records: parse_arrival(arrival, namespace_dictionary, location, cursor)

    departure_records = record.findall("namespace:dep", {"namespace": namespace_dictionary.get("fc")})
    for departure in departure_records: parse_departure(departure, namespace_dictionary, location, cursor)

    passing_records = record.findall("namespace:pass", {"namespace": namespace_dictionary.get("fc")})
    for passing in passing_records: parse_passing(passing, namespace_dictionary, location, cursor)

    # departure_records = record.findall("namespace:dep", {"namespace": namespace_dictionary.get("fc")})
    # for departure in departure_records: parse_train_status_location_departure_record(departure, namespace_dictionary, train_status, train_status_location)

    # passing_records = record.findall("namespace:pass", {"namespace": namespace_dictionary.get("fc")})
    # for passing in passing_records: parse_train_status_location_passing_record(passing, namespace_dictionary, train_status, train_status_location)

    # # platform_records = record.findall("namespace:plat", {"namespace": namespace_dictionary.get("fc")})
    # # for platform in platform_records: parse_train_status_location_platform_record(platform, namespace_dictionary, train_status, train_status_location)

    # suppressed_records = record.findall("namespace:suppr", {"namespace": namespace_dictionary.get("fc")})
    # for suppressed in suppressed_records: parse_train_status_location_suppressed_record(suppressed, namespace_dictionary, train_status, train_status_location)

    # # length_records = record.findall("namespace:length", {"namespace": namespace_dictionary.get("fc")})
    # # for length in length_records: parse_train_status_location_length_record(length, namespace_dictionary, train_status, train_status_location)

    # # detach_records = record.findall("namespace:detachFront", {"namespace": namespace_dictionary.get("fc")})
    # # for detach in detach_records: parse_train_status_location_detach_record(detach, namespace_dictionary, train_status, train_status_location)

def parse_train_status(record: ET.Element, namespace_dictionary: dict[str, str], date: datetime.date, update: collection.pushport.classes.Update, cursor):
    train_status = collection.pushport.classes.TrainStatus(
        update = update,
        rid = record.get("rid"),
        trainId = record.get("uid"),
        startDate = parse_date(record.get("ssd")),
        isReverse = parse_boolean(record.get("isCharter"), False)
    )

    if train_status.startDate != date: return
    # reason_record = record.find("namespace:LateReason", {"namespace": namespace_dictionary.get("fc")})
    # reason = None
    # if reason_record != None: reason = reason_record.text

    # if reason != None:
    #     cursor.execute(
    #     """
    #     UPDATE stops SET reason = ? WHERE rid = ?
    #     """,
    #     (reason, rtti_id))
    #     connection.commit()

    # train_status = TrainStatus(rtti_id, train_uid, scheduled_start_date, reason)

    location_records = record.findall("namespace:Location", {"namespace": namespace_dictionary.get("fc")})
    for location in location_records: parse_location(location, namespace_dictionary, train_status, cursor)

def parse_update(record: ET.Element, namespace_dictionary: dict[str, str], date: datetime.date, cursor):
    update = collection.pushport.classes.Update(
        origin = record.get("updateOrigin"),
        source = record.get("requestSource"),
        id = record.get("requestID"),
    )

    train_status_records = record.findall("namespace:TS", {"namespace": namespace_dictionary.get("")})
    for train_status in train_status_records: parse_train_status(train_status, namespace_dictionary, date, update, cursor)

def replay_file(path, date: datetime.date):
    cursor = Database.cursor()

    chunk_count = 0
    print(chunk_count)

    for chunk_data in chunk_file(path):
        chunk_data = chunk_data.strip()
        if not chunk_data:
            continue

        # Get the chunks data
        chunk = chunk_data.split("<?xml version=\"1.0\" encoding=\"utf-8\"?>", 1)[1]
        namespace_dictionary = collection.xml.fetch_namespace_dictionary(chunk)

        # Parse the chunk document
        root = ET.fromstring(chunk)

        update_records = root.findall("namespace:uR", {"namespace": namespace_dictionary.get("")})
        for update in update_records: parse_update(update, namespace_dictionary, date, cursor)

        chunk_count += 1
        print(chunk_count)

    Database.commit()

current_date = start_date

while current_date <= end_date:
    print(current_date)
    for hour in range (0,24):
        path = f"{darwin_location}/{current_date.strftime("%Y-%m-%d")}-{hour:02}.pport"
        replay_file(path, current_date)

    current_date += datetime.timedelta(days=1)

Database.close()