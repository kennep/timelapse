#!/bin/sh -e
docker run -d --rm -it --env-file test.env -P --name timelapse-server timelapse
port=`docker port timelapse-server 8080 | cut -d: -f2`
echo "Timelapse Server URL: http://localhost:$port/"
docker attach timelapse-server
