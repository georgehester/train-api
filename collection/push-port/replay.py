import sqlite3
import json
import datetime
import xml.etree.ElementTree as ET
import io
import os
from dataclasses import dataclass

@dataclass
class TrainStatus:
    rtti_id: str
    train_uid: str
    scheduled_start_date: str
    reason: str

@dataclass
class TrainStatusLocation:
    tiploc: str

darwin_location =  "/Volumes/LaCie/train/darwin"

start_date = datetime.date(2026, 1, 31)
end_date = datetime.date(2026, 2, 1)

train_uids = ["Y25969"]
rids = ["202601318925969"]
date_check = datetime.date(2026, 1, 31)

connection = sqlite3.connect("train.db")
cursor = connection.cursor()

cursor.execute(
"""
SELECT tiploc FROM stations
"""
)
station_tiplocs = [row[0] for row in cursor.fetchall()]

def parse_time(time):
    if time == None:
        return "XX-XX"
    split = time.split(":")
    return f"{split[0]}-{split[1]}"

def fetch_namespace_dictionary(chunk):
    output = {}

    for _, element in ET.iterparse(io.StringIO(chunk), events=("start-ns",)):
        prefix, uri = element
        output[prefix] = uri

    return output

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

@dataclass
class TrainStatusTimeData:
    estimated_time: str
    working_estimated_time: str
    actual_time: str
    actual_removed: str
    actual_class: str
    estimated_minimum: str
    estimated_unknown: str
    delayed: str
    actual_source: str
    actual_source_instance: str

def parse_train_status_time_data(record: ET.Element) -> TrainStatusTimeData:
    return TrainStatusTimeData(
        estimated_time = record.get("et"),
        working_estimated_time = record.get("wet"),
        actual_time = record.get("at"),
        actual_removed = record.get("atRemoved"), # Replaced by estimated
        actual_class = record.get("atClass"), # Class of actual time
        estimated_minimum = record.get("etmin"), # Lower limit for estimated time
        estimated_unknown = record.get("etUnknown"), # Unknown estimated time
        delayed = record.get("delayed"), # Delay unknown
        actual_source = record.get("src"), # Source of the actual time
        actual_source_instance = record.get("srcInst")
    )

def print_train_status_time_data(data: TrainStatusTimeData):
    print(f"{"Estimated Time":<48} - {data.estimated_time}")
    print(f"{"Working Estimated Time":<48} - {data.working_estimated_time}")
    print(f"{"Actual Time":<48} - {data.actual_time}")
    print(f"{"Actual Removed":<48} - {data.actual_removed}")
    print(f"{"Actual Class":<48} - {data.actual_class}")
    print(f"{"Estimated Minimum":<48} - {data.estimated_minimum}")
    print(f"{"Estimated Unknown":<48} - {data.estimated_unknown}")
    print(f"{"Delayed":<48} - {data.delayed}")
    print(f"{"Actual Source":<48} - {data.actual_source}")
    print(f"{"Actual Source Instance":<48} - {data.actual_source_instance}")


def parse_train_status_location_arrival_record(record: ET.Element, namespace_dictionary: dict[str, str], train_status: TrainStatus, train_status_location: TrainStatusLocation):
    print()
    print("ARRIVAL")
    print(f"{"Station":<48} - {train_status_location.tiploc}")
    data = parse_train_status_time_data(record)
    print_train_status_time_data(data)
    print(f"{"Reason":<48} - {train_status.reason}")

    if data.actual_removed == "true":
        cursor.execute(
        """
        UPDATE stops SET actual_arrival = NULL WHERE tiploc = ? AND rid = ?
        """,
        (train_status_location.tiploc, train_status.rtti_id))
        connection.commit()

    if data.actual_time != None:
        cursor.execute(
        """
        UPDATE stops SET actual_arrival = ? WHERE tiploc = ? AND rid = ?
        """,
        (parse_time(data.actual_time), train_status_location.tiploc, train_status.rtti_id))
        connection.commit()


def parse_train_status_location_departure_record(record: ET.Element, namespace_dictionary: dict[str, str], train_status: TrainStatus, train_status_location: TrainStatusLocation):
    print()
    print("DEPARTURE")
    print(f"{"Station":<48} - {train_status_location.tiploc}")
    data = parse_train_status_time_data(record)
    print_train_status_time_data(data)
    print(f"{"Reason":<48} - {train_status.reason}")

    if data.actual_removed == "true":
        cursor.execute(
        """
        UPDATE stops SET actual_departure = NULL WHERE tiploc = ? AND rid = ?
        """,
        (train_status_location.tiploc, train_status.rtti_id))
        connection.commit()

    if data.actual_time != None:
        cursor.execute(
        """
        UPDATE stops SET actual_departure = ? WHERE tiploc = ? AND rid = ?
        """,
        (parse_time(data.actual_time), train_status_location.tiploc, train_status.rtti_id))
        connection.commit()

def parse_train_status_location_passing_record(record: ET.Element, namespace_dictionary: dict[str, str], train_status: TrainStatus, train_status_location: TrainStatusLocation):
    print()
    print("PASSING")
    print(f"{"Station":<48} - {train_status_location.tiploc}")
    data = parse_train_status_time_data(record)
    print_train_status_time_data(data)
    print(f"{"Reason":<48} - {train_status.reason}")

    if data.actual_removed == "true":
        cursor.execute(
        """
        UPDATE stops SET actual_passing = NULL WHERE tiploc = ? AND rid = ?
        """,
        (train_status_location.tiploc, train_status.rtti_id))
        connection.commit()

    if data.actual_time != None:
        cursor.execute(
        """
        UPDATE stops SET actual_passing = ? WHERE tiploc = ? AND rid = ?
        """,
        (parse_time(data.actual_time), train_status_location.tiploc, train_status.rtti_id))
        connection.commit()

def parse_train_status_location_suppressed_record(record: ET.Element, namespace_dictionary: dict[str, str], train_status: TrainStatus, train_status_location: TrainStatusLocation):
    print()
    print("SUPPRESSED")
    print(f"{"Station":<48} - {train_status_location.tiploc}")
    print(f"{"Suppressed":<48} - {record.text}")

def parse_train_status_location_record(record: ET.Element, namespace_dictionary: dict[str, str], train_status: TrainStatus):
    tiploc = record.get("tpl")

    train_status_location = TrainStatusLocation(tiploc)

    arrival_records = record.findall("namespace:arr", {"namespace": namespace_dictionary.get("fc")})
    for arrival in arrival_records: parse_train_status_location_arrival_record(arrival, namespace_dictionary, train_status, train_status_location)

    departure_records = record.findall("namespace:dep", {"namespace": namespace_dictionary.get("fc")})
    for departure in departure_records: parse_train_status_location_departure_record(departure, namespace_dictionary, train_status, train_status_location)

    passing_records = record.findall("namespace:pass", {"namespace": namespace_dictionary.get("fc")})
    for passing in passing_records: parse_train_status_location_passing_record(passing, namespace_dictionary, train_status, train_status_location)

    # platform_records = record.findall("namespace:plat", {"namespace": namespace_dictionary.get("fc")})
    # for platform in platform_records: parse_train_status_location_platform_record(platform, namespace_dictionary, train_status, train_status_location)

    suppressed_records = record.findall("namespace:suppr", {"namespace": namespace_dictionary.get("fc")})
    for suppressed in suppressed_records: parse_train_status_location_suppressed_record(suppressed, namespace_dictionary, train_status, train_status_location)

    # length_records = record.findall("namespace:length", {"namespace": namespace_dictionary.get("fc")})
    # for length in length_records: parse_train_status_location_length_record(length, namespace_dictionary, train_status, train_status_location)

    # detach_records = record.findall("namespace:detachFront", {"namespace": namespace_dictionary.get("fc")})
    # for detach in detach_records: parse_train_status_location_detach_record(detach, namespace_dictionary, train_status, train_status_location)

def parse_train_status_record(record: ET.Element, namespace_dictionary: dict[str, str]):
    rtti_id = record.get("rid") # Maximum length 16
    train_uid = record.get("uid") # Length 6
    scheduled_start_date = record.get("ssd")

    if rtti_id not in rids: return

    reason_record = record.find("namespace:LateReason", {"namespace": namespace_dictionary.get("fc")})
    reason = None
    if reason_record != None: reason = reason_record.text

    if reason != None:
        cursor.execute(
        """
        UPDATE stops SET reason = ? WHERE rid = ?
        """,
        (reason, rtti_id))
        connection.commit()

    train_status = TrainStatus(rtti_id, train_uid, scheduled_start_date, reason)

    location_records = record.findall("namespace:Location", {"namespace": namespace_dictionary.get("fc")})
    for location in location_records: parse_train_status_location_record(location, namespace_dictionary, train_status)

def replay_file(path):
    for chunk_data in chunk_file(path):
        if not chunk_data.strip(): continue

        # Get the chunks data
        chunk = chunk_data.strip().split("<?xml version=\"1.0\" encoding=\"utf-8\"?>", 1)[1]
        namespace_dictionary = fetch_namespace_dictionary(chunk)

        # Parse the chunk document
        root = ET.fromstring(chunk)
        # status = root.find("namespace:FailureResp", {"namespace": root.get("xmlns")})

        updates = root.findall("namespace:uR", {"namespace": namespace_dictionary.get("")})

        for update in updates:
            train_status_records = update.findall("namespace:TS", {"namespace": namespace_dictionary.get("")})

            for record in train_status_records: parse_train_status_record(record, namespace_dictionary)








current_date = start_date

while current_date <= end_date:
    for hour in range (0,24):
        path = f"{darwin_location}/{current_date.strftime("%Y-%m-%d")}-{hour:02}.pport"
        replay_file(path)

    current_date += datetime.timedelta(days=1)


# replay_file("2026-01-01-00.pport")


connection.close()