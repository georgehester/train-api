from dataclasses import dataclass

@dataclass
class Update:
    origin: str # default none optional
    source: str # default none optional
    id: str # default none optional

@dataclass
class TrainStatus:
    update: Update
    rid: str
    trainId: str
    startDate: str
    isReverse: bool # default false optional

@dataclass
class Location:
    trainStatus: TrainStatus
    locationTIPLOC: str
    # workingTimeArrival: str # optional
    # workingTimeDeparture: str # optional
    # workingTimePassing: str # optional
    # publicTimeArrival: str # optional
    # publicTimeDeparture: str # optional

@dataclass
class Time:
    location: Location
    estimatedTime: str # optional
    workingEstimatedTime: str # optional
    actualTime: str # optional
    actualRemoved: bool # default false optional
    actualClass: str # optional
    estimatedMinimum: str # optional
    estimatedUnknown: str # default false optional
    isDelayed: str # default false optional
    source: str # optional
    sourceInstance: str # optional