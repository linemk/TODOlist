# ЭТАП СБОРКИ (builder)
FROM golang:1.23 AS builder

# Сборка под ARM64 (M1)
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=arm64

WORKDIR /app

# копируем файлы для зависимостей
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /my_app ./app/cmd

FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /my_app .

COPY web ./web

RUN mkdir -p /app/database

EXPOSE 7540

CMD ["/app/my_app"]