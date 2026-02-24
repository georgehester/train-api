from dataclasses import dataclass

@dataclass
class Timetable:
    id: str

@dataclass
class Journey:
    timetable: Timetable
    rid: str
    trainId: str
    trainHeadcode: str
    startDate: str
    trainOperatorCode: str
    status: str # default P optional
    trainCategory: str # default OO optional
    isPassengerService: bool # default true optional
    isDeleted: bool # default false optional
    isCharter: bool # default false optional

@dataclass
class Origin:
    journey: Journey
    workingTimeArrival: str # optional
    workingTimeDeparture: str
    falseDestination: str # optional
    locationTIPLOC: str
    activityCodes: str # optional default "  "
    plannedActivityCodes: str # optional
    cancelled: bool # default false optional
    platform: str # optional
    publicTimeArrival: str # optional
    publicTimeDeparture: str # optional

@dataclass
class OperationalOrigin:
    journey: Journey
    workingTimeArrival: str # optional
    workingTimeDeparture: str
    locationTIPLOC: str
    activityCodes: str # optional default "  "
    plannedActivityCodes: str # optional
    cancelled: bool # default false optional
    platform: str # optional

@dataclass
class Intermediate:
    journey: Journey
    workingTimeArrival: str # optional
    workingTimeDeparture: str
    falseDestination: str # optional
    routeDelay: str # default 0 optional
    locationTIPLOC: str
    activityCodes: str # optional default "  "
    plannedActivityCodes: str # optional
    cancelled: bool # default false optional
    platform: str # optional
    publicTimeArrival: str # optional
    publicTimeDeparture: str # optional

@dataclass
class OperationalIntermediate:
    journey: Journey
    workingTimeArrival: str # optional
    workingTimeDeparture: str
    routeDelay: str # default 0 optional
    locationTIPLOC: str
    activityCodes: str # optional default "  "
    plannedActivityCodes: str # optional
    cancelled: bool # default false optional
    platform: str # optional

@dataclass
class Passing:
    journey: Journey
    workingTimePassing: str
    routeDelay: str # default 0 optional
    locationTIPLOC: str
    activityCodes: str # optional default "  "
    plannedActivityCodes: str # optional
    cancelled: bool # default false optional
    platform: str # optional

@dataclass
class Destination:
    journey: Journey
    workingTimeArrival: str
    workingTimeDeparture: str # optional
    routeDelay: str # default 0 optional
    locationTIPLOC: str
    activityCodes: str # optional default "  "
    plannedActivityCodes: str # optional
    cancelled: bool # default false optional
    platform: str # optional
    publicTimeArrival: str # optional
    publicTimeDeparture: str # optional

@dataclass
class OperationalDestination:
    journey: Journey
    workingTimeArrival: str
    workingTimeDeparture: str # optional
    routeDelay: str # default 0 optional
    locationTIPLOC: str
    activityCodes: str # optional default "  "
    plannedActivityCodes: str # optional
    cancelled: bool # default false optional
    platform: str # optional