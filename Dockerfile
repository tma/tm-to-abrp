FROM golang:1.18.0-alpine3.15 AS builder

RUN apk add --no-cache git

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -o /go/bin/main -v cmd/main.go

# ----------------

FROM alpine:3.15.3

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /go/bin/main /app/cmd/main
COPY --from=builder /go/src/app/web /app/web

ENTRYPOINT /app/cmd/main
EXPOSE 3000
