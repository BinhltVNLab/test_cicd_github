ARG GO_VERSION=1.13.1

FROM golang:${GO_VERSION}-alpine as builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

RUN cd /app/cmd/frontapi && \ 
  CGO_ENABLED=0 go build -installsuffix 'static' -o /frontapi

FROM scratch AS final

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/group /etc/

COPY --from=builder /frontapi /frontapi

EXPOSE 8080

USER appuser:appuser

ENTRYPOINT ["/frontapi"]