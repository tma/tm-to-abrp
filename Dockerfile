FROM --platform=$BUILDPLATFORM golang:1.20.0-bullseye AS builder-base

ARG TARGETOS TARGETARCH

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    apt-get install -y git ca-certificates tzdata && \
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

# ----------------

FROM builder-base AS builder-test

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build \
    -o /go/bin/main -v cmd/main.go

# ----------------

FROM builder-base AS builder-release

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build \
    -ldflags="-w -s" -a -installsuffix cgo \
    -o /go/bin/main -v cmd/main.go

# ----------------

FROM scratch AS base

COPY --from=builder-base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder-base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder-base /etc/passwd /etc/passwd
COPY --from=builder-base /etc/group /etc/group

COPY --from=builder-base /go/src/app/web /app/web

# ----------------

FROM base AS release

COPY --from=builder-release /go/bin/main /app/cmd/main

WORKDIR /app
ENTRYPOINT [ "/app/cmd/main" ]
EXPOSE 3000

# ----------------

FROM debian:bullseye AS test

COPY --from=base / /
COPY --from=builder-test /go/bin/main /app/cmd/main

WORKDIR /app
ENTRYPOINT [ "/app/cmd/main" ]
EXPOSE 3000

# ---------------- default target

FROM test
