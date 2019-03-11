#!/bin/sh -e
port=`docker port timelapse-server 8080 | cut -d: -f2`
if test -z "$port"; then
    echo "Local timelapse server not started, please do ./run-server.sh first!"
    exit 1
fi
(cd cmd/timelapse && go build -o ../../build/timelapse) && build/timelapse --server-url http://localhost:$port "$@"