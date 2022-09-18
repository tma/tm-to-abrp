FROM golang:1.19.1-alpine3.16 AS builder

ARG BUILD_OS=${TARGETOS:-linux}
ARG BUILD_ARCH=${TARGETARCH:-amd64}

RUN apk update && \
  apk add --no-cache git ca-certificates tzdata && \
  update-ca-certificates

ENV USER=application
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go mod verify

RUN GOOS=$BUILD_OS GOARCH=$BUILD_ARCH go build -ldflags="-w -s" -o /go/bin/main -v cmd/main.go

# ----------------

FROM scratch

USER application:application

WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/main /app/cmd/main
COPY --from=builder /go/src/app/web /app/web

ENTRYPOINT /app/cmd/main
EXPOSE 3000
