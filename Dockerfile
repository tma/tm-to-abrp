FROM golang:1.17-alpine3.14 AS builder

RUN apk add --no-cache git

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -o /go/bin/app -v ./...

# ----------------

FROM alpine:3.14

RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app

ENTRYPOINT /app
EXPOSE 3000
