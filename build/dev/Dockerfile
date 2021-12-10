ARG GO_VERSION=1.13.1

FROM golang:${GO_VERSION}-alpine as builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
RUN go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

ENTRYPOINT ["go", "run", "/app/cmd/frontapi/main.go"]