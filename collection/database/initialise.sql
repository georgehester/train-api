CREATE TABLE IF NOT EXISTS stations (
    tiploc VARCHAR(7) PRIMARY KEY,
    nlc TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    crs VARCHAR(3) UNIQUE NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS journeys (
    id TEXT PRIMARY KEY,
    train_operator_code TEXT NOT NULL,
    train_category TEXT DEFAULT 'OO',
    passenger_service BOOLEAN DEFAULT TRUE,
    deleted BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS stops (
    id SERIAL PRIMARY KEY,
    journey_id TEXT NOT NULL,
    stop_type VARCHAR(4) NOT NULL,
    station_tiploc VARCHAR(7) NOT NULL,
    working_arrival TIME,
    working_departure TIME,
    working_passing TIME,
    public_arrival TIME,
    public_departure TIME,
    public_passing TIME,
    cancelled BOOLEAN DEFAULT FALSE,
    platform VARCHAR(3),
    actual_arrival TIME,
    actual_departure TIME,
    actual_passing TIME,
    FOREIGN KEY(journey_id) REFERENCES journeys(id),
    FOREIGN KEY(station_tiploc) REFERENCES stations(tiploc)
);

CREATE TABLE IF NOT EXISTS stations_analysis (
    tiploc VARCHAR(7) PRIMARY KEY,
    service_count INTEGER NOT NULL,
    delay_average_commute REAL NOT NULL,
    delay_rank_commute INTEGER NOT NULL,
    delay_average REAL NOT NULL,
    delay_rank INTEGER NOT NULL,
    FOREIGN KEY(tiploc) REFERENCES stations(tiploc)
);

-- CREATE INDEX IF NOT EXISTS idx_stops_station_tiploc_journey_id
-- ON stops (station_tiploc, journey_id);

CREATE TABLE IF NOT EXISTS administrators (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    forename TEXT NOT NULL,
    surname TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS customers (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    forename TEXT NOT NULL,
    surname TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS applications (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    key TEXT NOT NULL,
    customer_id TEXT NOT NULL,
    approved BOOLEAN DEFAULT FALSE,
    FOREIGN KEY(customer_id) REFERENCES customers(id)
);