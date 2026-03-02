#!/usr/bin/env bash

docker exec -t database pg_dump -U application -d train -F c -f /tmp/backup.dump
docker cp database:/tmp/backup.dump /Volumes/LaCie/database/backup.dump