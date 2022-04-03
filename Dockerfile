FROM golang:1.18.0-alpine3.15 AS builder

RUN apk add --no-cache git

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -o /go/bin/main -v ./...

# ----------------

FROM alpine:3.15.3

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /go/bin/main /app/main
COPY --from=builder /go/src/app/templates /app/templates
COPY --from=builder /go/src/app/public /app/public

ENTRYPOINT /app/main
EXPOSE 3000
