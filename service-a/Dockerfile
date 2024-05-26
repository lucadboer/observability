FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o service-a .

FROM alpine

WORKDIR /app

EXPOSE 8080

COPY --from=builder /app/service-a .

ENTRYPOINT ["./service-a"]
