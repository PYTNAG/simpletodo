# Build stage
FROM golang:1.21-alpine3.18 AS builder
WORKDIR /app
COPY . .

RUN go build -o main main.go

RUN apk add curl

RUN curl -O -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz
RUN tar xvzf migrate.linux-amd64.tar.gz migrate
RUN rm migrate.linux-amd64.tar.gz

# Run stage
FROM alpine:3.18
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

COPY /db/migration ./migration 

COPY app.env .

COPY start.sh .
COPY wait-for.sh . 

EXPOSE 8090
CMD [ "/app/main" ]

ENTRYPOINT [ "/app/start.sh" ]