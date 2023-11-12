# Build stage
FROM golang:1.21-alpine3.18 AS builder
WORKDIR /app
COPY . .

RUN go build -o main main.go

# Run stage
FROM alpine:3.18
WORKDIR /app

COPY --from=builder /app/main .

COPY /db/migration ./db/migration 

COPY app.env .

COPY start.sh .
COPY wait-for.sh . 

EXPOSE 8090
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]