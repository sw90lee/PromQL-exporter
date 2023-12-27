FROM golang:alpine as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
WORKDIR /usr/src/app
COPY . .

ENV GO111MODULE=on

#RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-s -w" -o bin/main ./cmd/exporter.go


### Executable Image
FROM alpine

COPY config.yml .

COPY cnf_config.yml .

COPY app_config.yml .

COPY --from=builder /usr/src/app/bin/main ./exporter

EXPOSE 8080

ENTRYPOINT ["./exporter"]