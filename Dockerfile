FROM golang:1.18 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./

COPY . .

RUN go mod download && \
    go mod verify && \
    go test -v ./... -coverprofile cover.out && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
 
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

COPY --from=builder /usr/src/app/app ./

CMD ["./app"]
