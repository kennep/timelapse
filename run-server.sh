#!/bin/sh -e
docker_image=timelapse
docker_args=
if test "$1" = "--dev"; then
    docker_image=timelapse-dev
    docker_args="-v $(realpath $(dirname $0))/endpoints/static:/static"
fi
docker run -d --rm -it --env-file test.env -p 5000:8080 --name timelapse-server $docker_args $docker_image
port=`docker port timelapse-server 8080 | cut -d: -f2`
echo "Timelapse Server URL: http://localhost:$port/"
docker attach timelapse-server
