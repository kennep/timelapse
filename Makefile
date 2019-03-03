.PHONY: build run-server clean

build:
	docker build . -t timelapse
	test -d ./build || mkdir ./build
	docker create --name timelapse timelapse
	docker cp timelapse:/bin/timelapse-server ./build/
	docker cp timelapse:/bin/timelapse ./build/
	docker rm timelapse

run-server: build-server
	./run-server.sh

clean:
	rm -f ./cmd/timelapse-server/timelapse-server
	rm -f ./cmd/timelapse/timelapse
	rm -Rf ./build
