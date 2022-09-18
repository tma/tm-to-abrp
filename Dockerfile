FROM golang:1.19.1-bullseye AS builder

ARG TARGETOS
ARG TARGETARCH

RUN apt-get update && \
    apt-get install git ca-certificates tzdata && \
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

WORKDIR $GOPATH/src/app

COPY go.mod .
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags="-w -s" -a -installsuffix cgo \
    -o /go/bin/main -v cmd/main.go

# ----------------

FROM scratch

USER application:application

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/main /app/cmd/main
COPY --from=builder /go/src/app/web /app/web

WORKDIR /app

ENTRYPOINT [ "/app/cmd/main" ]
EXPOSE 3000
