FROM golang:1.11-stretch as builder

WORKDIR /app

COPY go.mod /app
COPY go.sum /app

RUN go mod download

COPY . /app

ARG GO_BUILD_ARGS
ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go generate ./...
RUN for c in `find cmd -mindepth 1 -maxdepth 1 -type d`; do (cd $c && go build $GO_BUILD_ARGS); done

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/cmd/timelapse-server/timelapse-server /bin/timelapse-server
COPY --from=builder /app/cmd/timelapse/timelapse /bin/timelapse

EXPOSE 8080

WORKDIR /
ENTRYPOINT [ "/bin/timelapse-server" ]
