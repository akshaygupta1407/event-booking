FROM golang:1.25-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /event-booking ./cmd/api

FROM alpine:3.21
WORKDIR /app
COPY --from=build /event-booking /usr/local/bin/event-booking

EXPOSE 8080
CMD ["event-booking"]
