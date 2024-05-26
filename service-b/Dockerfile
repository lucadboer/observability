FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o service-b .

FROM alpine

WORKDIR /app

EXPOSE 8081

COPY --from=builder /app/service-b .
COPY .env .   

ENTRYPOINT ["./service-b"]
