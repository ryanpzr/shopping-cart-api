FROM golang:1.24.3 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o app ./cmd && go build -o activity-consumer ./cmd/activity-consumer

# The activity consumer: Kafka -> MongoDB (docker compose --profile activity-log).
FROM debian:bullseye-slim AS consumer
WORKDIR /app
COPY --from=build /app/activity-consumer .
CMD ["./activity-consumer"]

# The API. It is the last stage, so a plain `docker build .` produces it.
FROM debian:bullseye-slim AS api
WORKDIR /app
COPY --from=build /app/app .
EXPOSE 8080
CMD ["./app"]
