PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS stations (
    tiploc TEXT PRIMARY KEY,
    nlc TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    crs TEXT UNIQUE NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS journeys (
    id TEXT PRIMARY KEY,
    train_operator_code TEXT NOT NULL,
    train_category TEXT DEFAULT 'OO',
    passenger_service INTEGER DEFAULT 1,
    deleted INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stops (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    journey_id TEXT NOT NULL,
    stop_type TEXT NOT NULL,
    station_tiploc TEXT NOT NULL,
    working_arrival TEXT,
    working_departure TEXT,
    working_passing TEXT,
    public_arrival TEXT,
    public_departure TEXT,
    public_passing TEXT,
    cancelled INTEGER DEFAULT 0,
    platform TEXT,
    actual_arrival TEXT,
    actual_departure TEXT,
    actual_passing TEXT,
    FOREIGN KEY(journey_id) REFERENCES journeys(id),
    FOREIGN KEY(station_tiploc) REFERENCES stations(tiploc)
);