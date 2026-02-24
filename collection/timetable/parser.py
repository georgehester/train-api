import datetime
import xml.etree.ElementTree as ET
import collection.timetable.classes
from collection.database import Database
from typing import List
from collection.parser import parse_date, parse_boolean

def parse_origin(record: ET.Element, tiplocs: List[str], journey: collection.timetable.classes.Journey):
    origin = collection.timetable.classes.Origin(
        journey = journey,
        workingTimeArrival = record.get("wta"),
        workingTimeDeparture = record.get("wtd"),
        falseDestination = record.get("fd"),
        locationTIPLOC = record.get("tpl"),
        activityCodes = record.get("act"),
        plannedActivityCodes = record.get("planAct") or "  ",
        cancelled = parse_boolean(record.get("can"), False),
        platform = record.get("plat"),
        publicTimeArrival = record.get("pta"),
        publicTimeDeparture = record.get("ptd"),
    )

    if origin.locationTIPLOC not in tiplocs: return

    Database.cursor().execute(
        """
        INSERT INTO
            stops (journey_id, stop_type, station_tiploc, working_arrival, working_departure, cancelled, platform, public_arrival, public_departure)
        VALUES
            (?, ?, ?, ?, ?, ?, ?, ?, ?);
        """,
        (
            origin.journey.rid,
            "OR",
            origin.locationTIPLOC,
            origin.workingTimeArrival,
            origin.workingTimeDeparture,
            origin.cancelled,
            origin.platform,
            origin.publicTimeArrival,
            origin.publicTimeDeparture
        ),
    )
    Database.commit()
    print("OR")

def parse_operational_origin(record: ET.Element, tiplocs: List[str], journey: collection.timetable.classes.Journey):
    operational_origin = collection.timetable.classes.OperationalOrigin(
        journey=journey,
        workingTimeArrival=record.get("wta"),
        workingTimeDeparture=record.get("wtd"),
        locationTIPLOC=record.get("tpl"),
        activityCodes=record.get("act"),
        plannedActivityCodes=record.get("planAct") or "  ",
        cancelled=parse_boolean(record.get("can"), False),
        platform=record.get("plat"),
    )

    if operational_origin.locationTIPLOC not in tiplocs: return

    Database.cursor().execute(
        """
        INSERT INTO
            stops (journey_id, stop_type, station_tiploc, working_arrival, working_departure, cancelled, platform)
        VALUES
            (?, ?, ?, ?, ?, ?, ?);
        """,
        (
            operational_origin.journey.rid,
            "OPOR",
            operational_origin.locationTIPLOC,
            operational_origin.workingTimeArrival,
            operational_origin.workingTimeDeparture,
            operational_origin.cancelled,
            operational_origin.platform
        ),
    )
    Database.commit()
    print("OPOR")

def parse_intermediate(record: ET.Element, tiplocs: List[str], journey: collection.timetable.classes.Journey):
    intermediate = collection.timetable.classes.Intermediate(
        journey=journey,
        workingTimeArrival=record.get("wta"),
        workingTimeDeparture=record.get("wtd"),
        routeDelay=record.get("rdelay"),
        falseDestination=record.get("fd"),
        locationTIPLOC=record.get("tpl"),
        activityCodes=record.get("act"),
        plannedActivityCodes=record.get("planAct") or "  ",
        cancelled=parse_boolean(record.get("can"), False),
        platform=record.get("plat"),
        publicTimeArrival=record.get("pta"),
        publicTimeDeparture=record.get("ptd"),
    )

    if intermediate.locationTIPLOC not in tiplocs: return

    Database.cursor().execute(
        """
        INSERT INTO
            stops (journey_id, stop_type, station_tiploc, working_arrival, working_departure, cancelled, platform, public_arrival, public_departure)
        VALUES
            (?, ?, ?, ?, ?, ?, ?, ?, ?);
        """,
        (
            intermediate.journey.rid,
            "IP",
            intermediate.locationTIPLOC,
            intermediate.workingTimeArrival,
            intermediate.workingTimeDeparture,
            intermediate.cancelled,
            intermediate.platform,
            intermediate.publicTimeArrival,
            intermediate.publicTimeDeparture
        ),
    )
    Database.commit()
    print("IP")

def parse_operational_intermediate(record: ET.Element, tiplocs: List[str], journey: collection.timetable.classes.Journey):
    operational_intermediate = collection.timetable.classes.OperationalIntermediate(
        journey=journey,
        workingTimeArrival=record.get("wta"),
        workingTimeDeparture=record.get("wtd"),
        routeDelay=int(record.get("rdelay") or 0),
        locationTIPLOC=record.get("tpl"),
        activityCodes=record.get("act"),
        plannedActivityCodes=record.get("planAct") or "  ",
        cancelled=parse_boolean(record.get("can"), False),
        platform=record.get("plat"),
    )

    if operational_intermediate.locationTIPLOC not in tiplocs: return

    Database.cursor().execute(
        """
        INSERT INTO
            stops (journey_id, stop_type, station_tiploc, working_arrival, working_departure, cancelled, platform)
        VALUES
            (?, ?, ?, ?, ?, ?, ?);
        """,
        (
            operational_intermediate.journey.rid,
            "OPIP",
            operational_intermediate.locationTIPLOC,
            operational_intermediate.workingTimeArrival,
            operational_intermediate.workingTimeDeparture,
            operational_intermediate.cancelled,
            operational_intermediate.platform
        ),
    )
    Database.commit()
    print("OPIP")

def parse_passing(record: ET.Element, tiplocs: List[str], journey: collection.timetable.classes.Journey):
    passing = collection.timetable.classes.Passing(
        journey=journey,
        workingTimePassing=record.get("wtp"),
        routeDelay=int(record.get("rdelay") or 0),
        locationTIPLOC=record.get("tpl"),
        activityCodes=record.get("act"),
        plannedActivityCodes=record.get("planAct") or "  ",
        cancelled=parse_boolean(record.get("can"), False),
        platform=record.get("plat"),
    )

    if passing.locationTIPLOC not in tiplocs: return

    Database.cursor().execute(
        """
        INSERT INTO
            stops (journey_id, stop_type, station_tiploc, working_passing, cancelled, platform)
        VALUES
            (?, ?, ?, ?, ?, ?);
        """,
        (
            passing.journey.rid,
            "PP",
            passing.locationTIPLOC,
            passing.workingTimePassing,
            passing.cancelled,
            passing.platform
        ),
    )
    Database.commit()
    print("PP")

def parse_destination(record: ET.Element, tiplocs: List[str], journey: collection.timetable.classes.Journey):
    destination = collection.timetable.classes.Destination(
        journey=journey,
        workingTimeArrival=record.get("wta"),
        workingTimeDeparture=record.get("wtd"),
        routeDelay=int(record.get("rdelay") or 0),
        locationTIPLOC=record.get("tpl"),
        activityCodes=record.get("act"),
        plannedActivityCodes=record.get("planAct") or "  ",
        cancelled=parse_boolean(record.get("can"), False),
        platform=record.get("plat"),
        publicTimeArrival=record.get("pta"),
        publicTimeDeparture=record.get("ptd"),
    )

    if destination.locationTIPLOC not in tiplocs: return

    Database.cursor().execute(
        """
        INSERT INTO
            stops (journey_id, stop_type, station_tiploc, working_arrival, working_departure, cancelled, platform, public_arrival, public_departure)
        VALUES
            (?, ?, ?, ?, ?, ?, ?, ?, ?);
        """,
        (
            destination.journey.rid,
            "DT",
            destination.locationTIPLOC,
            destination.workingTimeArrival,
            destination.workingTimeDeparture,
            destination.cancelled,
            destination.platform,
            destination.publicTimeArrival,
            destination.publicTimeDeparture
        ),
    )
    Database.commit()
    print("DT")

def parse_operational_destination(record: ET.Element, tiplocs: List[str], journey: collection.timetable.classes.Journey):
    operational_destination = collection.timetable.classes.OperationalDestination(
        journey=journey,
        workingTimeArrival=record.get("wta"),
        workingTimeDeparture=record.get("wtd"),
        routeDelay=int(record.get("rdelay") or 0),
        locationTIPLOC=record.get("tpl"),
        activityCodes=record.get("act"),
        plannedActivityCodes=record.get("planAct") or "  ",
        cancelled=parse_boolean(record.get("can"), False),
        platform=record.get("plat"),
    )

    if operational_destination.locationTIPLOC not in tiplocs: return

    Database.cursor().execute(
        """
        INSERT INTO
            stops (journey_id, stop_type, station_tiploc, working_arrival, working_departure, cancelled, platform)
        VALUES
            (?, ?, ?, ?, ?, ?, ?);
        """,
        (
            operational_destination.journey.rid,
            "OPDT",
            operational_destination.locationTIPLOC,
            operational_destination.workingTimeArrival,
            operational_destination.workingTimeDeparture,
            operational_destination.cancelled,
            operational_destination.platform
        ),
    )
    Database.commit()
    print("OPDT")

def parse_journey(record: ET.Element, namespace_dictionary: dict[str, str], date: datetime.date, tiplocs: List[str], timetable: collection.timetable.classes.Timetable):
    journey = collection.timetable.classes.Journey(
        timetable = timetable,
        rid = record.get("rid"),
        trainId = record.get("uid"),
        trainHeadcode = record.get("trainId"),
        startDate = parse_date(record.get("ssd")),
        trainOperatorCode = record.get("toc"),
        status = record.get("status") or "P",
        trainCategory = record.get("trainId") or "OO",
        isPassengerService = parse_boolean(record.get("isPassengerSvc"), True),
        isDeleted = parse_boolean(record.get("deleted"), False),
        isCharter = parse_boolean(record.get("isCharter"), False)
    )

    if journey.isPassengerService == False: return
    if journey.startDate != date: return

    Database.cursor().execute(
        """
        INSERT INTO
            journeys (id, train_operator_code, train_category, passenger_service, deleted)
        VALUES
            (?, ?, ?, ?, ?)
        ON CONFLICT (id)
        DO UPDATE SET
            train_operator_code = EXCLUDED.train_operator_code,
            train_category = EXCLUDED.train_category,
            passenger_service = EXCLUDED.passenger_service,
            deleted = EXCLUDED.deleted;
        """,
        (
            journey.rid,
            journey.trainOperatorCode,
            journey.trainCategory,
            journey.isPassengerService,
            journey.isDeleted
        ),
    )
    Database.commit()

    origin_records = record.findall("namespace:OR", {"namespace": namespace_dictionary.get("")})
    for origin in origin_records: parse_origin(origin, tiplocs, journey)

    operational_origin_records = record.findall("namespace:OPOR", {"namespace": namespace_dictionary.get("")})
    for operational_origin in operational_origin_records: parse_operational_origin(operational_origin, tiplocs, journey)

    intermediate_records = record.findall("namespace:IP", {"namespace": namespace_dictionary.get("")})
    for intermediate in intermediate_records: parse_intermediate(intermediate, tiplocs, journey)

    operational_intermediate_records = record.findall("namespace:OPIP", {"namespace": namespace_dictionary.get("")})
    for operational_intermediate in operational_intermediate_records: parse_operational_intermediate(operational_intermediate, tiplocs, journey)

    passing_records = record.findall("namespace:PP", {"namespace": namespace_dictionary.get("")})
    for passing in passing_records: parse_passing(passing, tiplocs, journey)

    destination_records = record.findall("namespace:DT", {"namespace": namespace_dictionary.get("")})
    for destination in destination_records: parse_destination(destination, tiplocs, journey)

    operational_destination_records = record.findall("namespace:OPDT", {"namespace": namespace_dictionary.get("")})
    for operational_destination in operational_destination_records: parse_operational_destination(operational_destination, tiplocs, journey)

def parse_timetable(record: ET.Element, namespace_dictionary: dict[str, str], date: datetime.date, tiplocs: List[str]):
    timetable = collection.timetable.classes.Timetable(id=record.get("timetableID"))

    journey_records = record.findall("namespace:Journey", {"namespace": namespace_dictionary.get("")})
    for journey in journey_records: parse_journey(journey, namespace_dictionary, date, tiplocs, timetable)