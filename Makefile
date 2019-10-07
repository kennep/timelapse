.PHONY: build run-server clean

build:
	docker build . -t timelapse
	test -d ./build || mkdir ./build
	docker create --name timelapse timelapse
	docker cp timelapse:/bin/timelapse-server ./build/
	docker cp timelapse:/bin/timelapse ./build/
	docker rm timelapse

build-dev:
	docker build . --build-arg "GO_BUILD_ARGS=-tags dev" -t timelapse-dev

run-server: build
	./run-server.sh

run-server-dev: build-dev
	./run-server.sh --dev

clean:
	rm -f ./cmd/timelapse-server/timelapse-server
	rm -f ./cmd/timelapse/timelapse
	rm -Rf ./build
