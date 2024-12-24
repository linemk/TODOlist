FROM golang:1.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o my_app ./app/cmd

FROM ubuntu:latest
WORKDIR /app


COPY --from=builder /app/web ./web
COPY --from=builder /app/my_app .

RUN mkdir -p /app/database

ENV TODO_PORT=7540
ENV TODO_DBFILE=./database/scheduler.db
ENV JWT_SECRET=secret_key
ENV TODO_PASSWORD=12345

EXPOSE 7540

CMD ["/app/my_app"]