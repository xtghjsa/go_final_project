# Этап сборки
FROM golang:1.22.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY web ./web
COPY api ./api

ENV TODO_DBFILE=./scheduler.db
ENV TODO_PORT=7540
ENV TODO_PASSWORD=8888

RUN CGO_ENABLED=0 GOOS=linux go build -o /my_app

CMD ["/my_app"]
