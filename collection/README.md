# Collection



> [!NOTE]
> You must be within the base folder to be able to run the python scripts.



## Stations

To generate the `stations` table in the database run the python script, make sure to modify the file path for your data source.

```sh
python -m collection.stations.generate`
```

## Timetable

To generate the `journeys` and base of the `stops` table run.

```sh
python -m collection.timetable.generate`
```

## Push Port

To replay the updates from the National Rail Darwin Push Port data run the python script.

```sh
python -m collection.pushport.generate`
```

## Stations Analysis

To generate the `stations_analysis` table run the following SQL script against the updated database.

```sql
INSERT INTO stations_analysis (
    tiploc,
    service_count,
    delay_average_commute,
    delay_rank_commute,
    delay_average,
    delay_rank
)
SELECT
    tiploc,
    service_count,
    COALESCE(delay_average_commute, 0) AS delay_average_commute,
    RANK() OVER (ORDER BY COALESCE(delay_average_commute, 0) DESC) AS delay_rank_commute,
    COALESCE(delay_average, 0) AS delay_average,
    RANK() OVER (ORDER BY COALESCE(delay_average, 0) DESC) AS delay_rank
FROM (
    SELECT
        stations.tiploc,
        COUNT(stops.id) AS service_count,
        AVG(
            CASE WHEN
                NOT (stops.working_arrival IS NULL AND stops.working_departure IS NULL)
                AND (
                    (stops.working_departure IS NOT NULL AND stops.actual_departure IS NOT NULL)
                    OR (stops.working_departure IS NULL AND stops.working_arrival IS NOT NULL
                    AND stops.actual_arrival IS NOT NULL)
                )
            THEN GREATEST(
                CASE
                    WHEN stops.working_departure IS NOT NULL
                        THEN EXTRACT(EPOCH FROM (stops.actual_departure - stops.working_departure))
                    ELSE
                        EXTRACT(EPOCH FROM (stops.actual_arrival - stops.working_arrival))
                END,
                0
            )
            ELSE NULL END
        ) / 60.0 AS delay_average,
        AVG(
            CASE WHEN
                NOT (stops.working_arrival IS NULL AND stops.working_departure IS NULL)
                AND (
                    (stops.working_departure IS NOT NULL AND stops.actual_departure IS NOT NULL AND stops.working_departure BETWEEN '08:00' AND '10:00')
                    OR (stops.working_departure IS NULL AND stops.working_arrival IS NOT NULL AND stops.actual_arrival IS NOT NULL AND stops.working_arrival BETWEEN '08:00' AND '10:00')
                )
            THEN GREATEST(
                CASE
                    WHEN stops.working_departure IS NOT NULL
                        THEN EXTRACT(EPOCH FROM (stops.actual_departure - stops.working_departure))
                    ELSE
                        EXTRACT(EPOCH FROM (stops.actual_arrival - stops.working_arrival))
                END,
                0
            )
            ELSE NULL END
        ) / 60.0 AS delay_average_commute
    FROM stations
    LEFT JOIN stops ON stops.station_tiploc = stations.tiploc
    GROUP BY stations.tiploc
) AS station_data;
```