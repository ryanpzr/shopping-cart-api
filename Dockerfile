FROM golang:1.24.3 AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o app ./cmd
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=build /app/app .
EXPOSE 8080
CMD ["./app"]